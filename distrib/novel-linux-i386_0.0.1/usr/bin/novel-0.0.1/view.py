#!/usr/bin/env python3
import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GObject

import json
import os
import sys
import logging
import log
from config import config

UP_KEY = 65362
DOWN_KEY = 65364
RIGHT_KEY = 65363
LEFT_KEY = 65361
SPACE_KEY = 32
HOME_KEY = 65360
END_KEY = 65367

class NovelWindow():
    NOVEL_DIR = config['novel_dirname']
    LOG_FILE = ".record.json"
    LOG_DIR = config['log_dirname']
    NOVEL_EXT = config['novel_extname']

    def __init__(self):
        builder = Gtk.Builder()
        builder.add_from_file("./novel.glade")
        self.win = builder.get_object("main_window")
        self.log_file_full_path = os.path.join(NovelWindow.LOG_DIR, NovelWindow.LOG_FILE) 

        self.treeview_menu = builder.get_object("treeview_menu")
        self.treeview_model = builder.get_object("liststore_menu")
        self.treeview_menu.set_model(self.treeview_model)

        self.treeview_selection = self.treeview_menu.get_selection()

        self.label_content = builder.get_object("label_content")
        self.content_adjustment = builder.get_object("viewport1").get_vadjustment()

        self.win.connect("delete-event", self.on_quit)
        builder.connect_signals(self)

        if len(sys.argv) < 2:
            print("Usage: %s 小说名称" %sys.argv[0])
            exit(1)

        self.read_novel(sys.argv[1])

        self._handle_novel_menus(self.novel["Menus"], 20)
        self.arange_novel_content()
        self.add_menulist()

        self._apply_custom_style(self.win)
        self._apply_custom_style(self.treeview_menu)
        self._apply_custom_style(self.label_content)


        self.init_vadjustment_val = None
        # 选中菜单的第一项
        self.treeview_selection.select_iter(self.treeview_model.get_iter(Gtk.TreePath(0)))

        self.win.maximize()

    def show(self):
        self.win.show_all()
        self.load_use_log() #只有当窗口显示的时候，滚动窗口的vadjustment的范围才会确定下来，这个时候，设置值才有用
        #GObject.timeout_add(500, self.on_timeout, None) #GTK内部计算滚动条的范围需要时间，所以这儿进行延时操作

    def read_novel(self, name):
        path = os.path.join(NovelWindow.NOVEL_DIR, name + NovelWindow.NOVEL_EXT)
        self.win.set_title(name)
        with open(path, 'rt') as f:
            self.novel = json.load(f)

    def add_menulist(self):
        # 添加内容
        for menu in self.novel["Menus"]:
            self.treeview_model.append([menu["Name"], menu["URL"]])
 
    def on_selection_changed(self, widget):
        model, iter = widget.get_selected()
        if model != None:
            path = model.get_path(iter)
                
            self.treeview_menu.scroll_to_cell(path, None, False, 0, 0) #可以保证选中的cell在当前视野中

            index = int(str(path))
            self.label_content.set_text(('\n%s\n\n' %self.novel["Chapters"][index]["Title"]) + 
                    self.novel["Chapters"][index]["Content"])
            self.content_adjustment.value_changed()

            if self.init_vadjustment_val != None:
                self.content_adjustment.set_upper(self.init_upper)
                self.content_adjustment.clamp_page(self.init_vadjustment_val,
                        self.init_vadjustment_val + self.content_adjustment.get_page_size())
                logging.debug("Now value is: %s" %self.content_adjustment.get_value())
                self.init_vadjustment_val = None
                self.init_upper = None
            else:
                self._home()

    def setup_text_tag(self):
        self.text_tag = self.textbuffer.create_tag(foreground="#555555",
                paragraph_background="#DFFAFF", size_points=19,
                family="方正启体简体"
                )

    def arange_novel_content(self):
        for chapter in self.novel["Chapters"]:
            if chapter == None:#防止某些网站上小说页面出现异常情况
                continue
            chapter["Content"] = self._handle_novel_content(chapter["Content"])

    def _handle_novel_content(self, content):
        contents = content.split('\n')
        contents = [e.strip() for e in contents]
        contents = ['\t%s' %e for e in contents if len(e) > 0]
        return '\n\n'.join(contents)

    def _handle_novel_menus(self, menus, maxlen):
        for menu in menus:
            if len(menu["Name"]) >= maxlen:
                menu["Name"] = menu["Name"][:maxlen - 3] + "..."


    def on_key_press(self, widget, key):
        flag, keyval = key.get_keyval()

        # 只捕获左右方向键，其余的抓取
        if flag and (keyval == LEFT_KEY or keyval == RIGHT_KEY or keyval == UP_KEY or keyval == DOWN_KEY):
            if keyval == UP_KEY:
                self._line_up()
            elif keyval == DOWN_KEY:
                self._line_down()
            return True
                
        return False

    def on_key_release(self, widget, key):
        flag, keyval = key.get_keyval()
        logging.debug("on key release, keyval: %s" %keyval)

        if flag:
            if keyval == LEFT_KEY:
                return self._prev_chapter()
            elif keyval == RIGHT_KEY:
                return self._next_chapter()
            elif keyval == SPACE_KEY:
                self._next_page()
        return True

    # 切换到下一章节
    def _next_chapter(self):
        model, iter = self.treeview_selection.get_selected()
        iter =  model.iter_next(iter)
        if iter != None:
            self.treeview_selection.select_iter(iter)
        return True

    # 切换到上一章节
    def _prev_chapter(self):
        model, iter = self.treeview_selection.get_selected()
        iter =  model.iter_previous(iter)
        if iter != None:
            self.treeview_selection.select_iter(iter)
        return True

    def _next_page(self):
        logging.debug("val: %g, page size: %g, upper: %g" %(self.content_adjustment.get_value(), 
            self.content_adjustment.get_page_size(), self.content_adjustment.get_upper()))
        self.content_adjustment.set_value(self.content_adjustment.get_value() 
                + self.content_adjustment.get_page_increment())

    def _apply_custom_style(self, widget):
        context = widget.get_style_context()
        css_provider = Gtk.CssProvider()
        if css_provider.load_from_path("./novel.css"):
            context.add_provider(css_provider, Gtk.STYLE_PROVIDER_PRIORITY_USER)

    def _home(self):
        logging.debug("scroll text view content to home")
        self.content_adjustment.set_value(0)

    def _line_up(self):
        logging.debug("move line to up")
        self.content_adjustment.set_value(self.content_adjustment.get_value() 
                - self.content_adjustment.get_step_increment())

    def _line_down(self):
        logging.debug("move line to down")
        self.content_adjustment.set_value(self.content_adjustment.get_value() 
                + self.content_adjustment.get_step_increment())

    def on_quit(self, widget, param):
        logging.debug("save record info on %s" %self.novel['Name'])
        self._load_log() #重新加载，防止在阅读期间有其他进程对该文件进行读写，即读取的内容是最新的

        # 每个记录，保存了当前正在阅读的章节索引，和当前滚动条的位置
        record = dict()
        record['Vadjustment'] = float('%.2f' %self.content_adjustment.get_value())
        record['Upper'] = float('%.2f' %self.content_adjustment.get_upper())
        
        model, iter = self.treeview_selection.get_selected()
        path = self.treeview_model.get_path(iter)
        index = int(str(path))
        record['LastChapterIndex'] = index 

        
        if self.records == None:
            self.records = dict()
        logging.debug("record [%s], [vajustment:%d], [chapter_index:%d]" %(self.novel['Name'], record['Vadjustment'], 
            record['LastChapterIndex']))
        self.records[self.novel["Name"]] = record
        with open(self.log_file_full_path, 'wt') as f:
            json.dump(self.records, f)
        
        Gtk.main_quit()

    def load_use_log(self):
        self._load_log()
        
        if self.records != None:
            self._use_log()

    def _load_log(self):
        logging.info("load log .....")
        if not os.path.isdir(NovelWindow.LOG_DIR):
            os.makedirs(NovelWindow.LOG_DIR)
        if os.path.isfile(self.log_file_full_path):
            with open(self.log_file_full_path, 'rt') as f:
                self.records = json.load(f)
        else:
            self.records = None

    def _use_log(self):
        if self.novel['Name'] in self.records:
            info = self.records[self.novel['Name']]
            logging.debug("log record:", info)

            # 设置menu的选中项
            self.init_vadjustment_val = float('%2.f' %info['Vadjustment'])
            self.init_upper = float('%.2f' %info['Upper'])
            #self.content_adjustment.set_value(self.init_vadjustment_val) #直接内容滚动条位置，因为有可能item的选中项没有发生变化
            self.content_adjustment.set_upper(self.init_upper)
            self.content_adjustment.clamp_page(self.init_vadjustment_val,
                    self.init_vadjustment_val + self.content_adjustment.get_page_size())

            logging.debug("Lower: %s, upper: %s, page_size: %s" %(self.content_adjustment.get_lower(), 
                self.content_adjustment.get_upper(),
                   self.content_adjustment.get_page_size()))
            self.treeview_selection.select_iter(self.treeview_model.get_iter(Gtk.TreePath(info['LastChapterIndex'])))   

    def on_timeout(self, params):
        self.load_use_log()
        return False

def main():
    log.config(sys.argv[2:])
    win = NovelWindow()
    win.show()

    Gtk.main()

if __name__ == "__main__":
    main()
