#! /usr/bin/env python3
import logging
import config

import gi
gi.require_version('Gtk', '3.0')
from gi.repository import Gtk

class DownloadWidget(Gtk.ScrolledWindow):
    """download widget represent download page
    """

    TEXT = 0
    PROGRESS = 1

    def __init__(self, model):
        logging.info("build download widget...")
        Gtk.ScrolledWindow.__init__(self, margin=20)

        self.model = model
        self._setup_filter()

        self.content_treeview = Gtk.TreeView.new_with_model(self.filter) #to use filter, must set filter as the model of treeview
        self.selection = self.content_treeview.get_selection()
        self._setup_columns()

        self.add(self.content_treeview)
        self.set_policy(Gtk.PolicyType.NEVER, Gtk.PolicyType.AUTOMATIC)

    def _setup_columns(self):
        """setup treeview columns
        """
        logging.info("setup columns of download widget")
        for i, title, kind in [(0, "小说名称", DownloadWidget.TEXT), (1, "作者名称", DownloadWidget.TEXT),
                (2, "下载进度", DownloadWidget.PROGRESS), (3, "操作类型", DownloadWidget.TEXT)]:
            if kind == DownloadWidget.TEXT:
                renderer = Gtk.CellRendererText()
                column = Gtk.TreeViewColumn(title, renderer, text=i)
            elif kind == DownloadWidget.PROGRESS:
                renderer = Gtk.CellRendererProgress()
                column = Gtk.TreeViewColumn(title, renderer, value=i)
            self.content_treeview.append_column(column)

    def _setup_filter(self):
        """setup the filter of list store
        """
        logging.info("setup the filter of list store")
        self.filter = self.model.filter_new()
        self.current_filter_keyword = None
        self.filter.set_visible_func(self._on_filter_func)

    def _on_filter_func(self, model, iter, data):
        """used to filter to showing content
        """
        if self.current_filter_keyword in (None, "下载"):
            return True
        
        if self.current_filter_keyword == "已经完成":
            return True if model[iter][2] == 100 else False

        if self.current_filter_keyword == "正在下载":
            return True if model[iter][2] >= 0 and model[iter][2] < 100 else False

        if self.current_filter_keyword == "下载操作":
            return True if model[iter][3] == "下载" else False

        if self.current_filter_keyword == "更新操作":
            return True if model[iter][3] == "更新" else False

    def switch(self, keyword):
        self.current_filter_keyword = keyword
        self.filter.refilter()


class SearchWidget:

    def __init__(self, builder, main_window):
        logging.info("from build extract search widget's controls...")

        self.widget = builder.get_object("search_widget")
        self._assert_control(self.widget, "search_widget")

        self.novel_img = builder.get_object("novel_img") 
        self._assert_control(self.novel_img, "novel_img")

        self.search_content = builder.get_object("search_content")
        self._assert_control(self.search_content, "search_content")

        self.novel_name = builder.get_object("novel_name")
        self._assert_control(self.novel_name, "novel_name")

        self.novel_author_name = builder.get_object("novel_author_name")
        self._assert_control(self.novel_author_name, "novel_author_name")

        self.novel_description = builder.get_object("novel_description")
        self._assert_control(self.novel_description, "novel_description")

        self.novel_last_chapter_name = builder.get_object("novel_last_chapter_name")
        self._assert_control(self.novel_last_chapter_name, "novel_last_chapter_name")

        self.search_btn = builder.get_object("search")
        self._assert_control(self.search_btn, "search")
        self.search_btn.connect("clicked", self._on_search)

        self.search_result_container = builder.get_object("search_result_container")
        self._assert_control(self.search_result_container, "search_result_container")

    def _assert_control(self, control, control_id):
        """assert control, ouput error msg
        """
        assert control, "can't get object '%s' from builder" %control_id

    def _on_search(self, button):
        """do search operation
        """
        self.search_result_container.show()

    def show(self):
        self.widget.show()

    def hide(self):
        self.widget.hide()

class MainWindow:

    def __init__(self):
        logging.info("parse glade file '%s'..." %'./app.glade')
        builder = Gtk.Builder()
        builder.add_from_file("./app.glade")
        builder.connect_signals(self)

        self._init_controls(builder)
        self._setup_menu()
    
    def _init_controls(self, builder):
        """"get objects from builder
        """
        logging.info("init controls from builder...")

        self.window = builder.get_object("main_window")
        self._assert_control(self.window, "main_window")


        self.menu = builder.get_object("menu")
        self._assert_control(self.menu, "menu")

        self.menu_selection = builder.get_object("menu_selection")
        self._assert_control(self.menu_selection, "menu_selection")

        self.menu_treestore = builder.get_object("menu_treestore")
        self._assert_control(self.menu_treestore, "menu_treestore")
        

        self.content_container = builder.get_object("content_container")
        self._assert_control(self.content_container, "content_container")

        self.search_widget = SearchWidget(builder, self)

        self.download_liststore = builder.get_object("download_liststore")
        self._assert_control(self.download_liststore, "download_liststore")

        self.download_widget = DownloadWidget(self.download_liststore)
        self.content_container.pack_start(self.download_widget, True, True, 0)

    def _setup_menu(self):
        """setup tree view menu
        """
        # add search menu item
        self.menu_treestore.append(None, ['搜索'])

        # add download menu item
        download_parent = self.menu_treestore.append(None, ['下载'])
        for item in ('正在下载', '已经完成', '更新操作', '下载操作'):
            self.menu_treestore.append(download_parent, [item])

    def _assert_control(self, control, control_id):
        """assert control, ouput error msg
        """
        assert control, "can't get object '%s' from builder" %control_id

    def show(self):
        """show all widget
        """
        self.window.show_all()

    def on_quit(self, widget):
        """listen destroy and exit full application
        """
        Gtk.main_quit()

    def on_menu_selection_changed(self, selection):
        """listen tree selection changed signal. Main to switch different page
        """
        model, iter = selection.get_selected()
        if iter != None:
            name = model[iter][0]
            if name == "搜索":
                self.search_widget.show()
                self.download_widget.hide()
            elif name in ('下载', '正在下载', '已经完成', '更新操作', '下载操作'):
                self.download_widget.show()
                self.search_widget.hide()
                self.download_widget.switch(name)

def main():
    config.config()
    win = MainWindow()
    win.show()

    Gtk.main()
        
if __name__ == '__main__':
    main()
