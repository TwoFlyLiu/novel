#! /usr/bin/env python3
import logging
import config
import sys
import os
import subprocess
import _thread
import threading
import json
import time

from collections import deque

import gi
gi.require_version('Gtk', '3.0')
from gi.repository import Gtk, GObject, Gdk
from gi.repository.GdkPixbuf import Pixbuf

THIS_SCRIPT_DIRNAME =  os.path.dirname(os.path.abspath(__file__))
SEARCH_EXECUTED_FILE = '../search/search'

class Work():

    def __init__(self, iter, index, value):
        self.iter = iter
        self.value = value
        self.index = index

class WorkQueue():
    """deque本身是线程安全的，所以没有必要再加锁"""

    def __init__(self):
        # self.lock = threading.Lock()
        self.deque = deque()

    def push(self, work):
        """入队
        """
        # self.lock.acquire()
        self.deque.append(work)
        # self.lock.release()

    def pop(self):
        """出队
        """
        # self.lock.acquire()
        work = self.deque.popleft()
        # self.lock.release()
        return work

    def is_empty(self):
        # self.lock.acquire()
        result = (len(self.deque) == 0)
        # self.lock.release()
        return result

class NovelsWidget(Gtk.Box):
    CONTEXT_MENU = """
        <ui>
            <popup name="PopupMenu">
                <menuitem action="view_novel"/>
                <menuitem action="remove_novel"/>
                <menuitem action="update_novel"/>
            </popup>
        </ui>
    """
    ICON_WIDTH = 100
    ICON_HEIGHT = 125

    def __init__(self, mgr):
        Gtk.Box.__init__(self, orientation=Gtk.Orientation.VERTICAL)
        self.mgr = mgr

        self.icon_view = Gtk.IconView(activate_on_single_click=False, item_width=100, item_padding=15)
        self.list_store = Gtk.ListStore(Pixbuf, str)
        self.icon_view.set_pixbuf_column(0)
        self.icon_view.set_text_column(1)
        self.icon_view.connect("item_activated", self.on_item_activated)

        self.icon_view.set_model(self._setup_filter())

        # 里面是假的数据
        files = [e for e in os.listdir("./json") if self._is_novel(e)]
        for  f in files:
            pixbuf = Pixbuf.new_from_file_at_scale("./json/%s.img" %f, NovelsWidget.ICON_WIDTH, 
                    NovelsWidget.ICON_HEIGHT, True)
            self.list_store.append([pixbuf, f])
        self.files = files

        sw = Gtk.ScrolledWindow()
        sw.set_policy(Gtk.PolicyType.NEVER, Gtk.PolicyType.AUTOMATIC)
        sw.add(self.icon_view)

        self.icon_view.connect("button-release-event", self.on_button_release)

        self.search_box = Gtk.Box(margin=10, margin_right=30)
        self.search_entry = Gtk.SearchEntry(width_request=300)
        self.search_entry.connect("search-changed", self.on_search_changed)

        self.search_box.pack_start(Gtk.Label(), True, True, 0)
        self.search_box.pack_start(self.search_entry, False, True, 0)

        self.pack_start(self.search_box, False, True, 0)
        self.pack_start(sw, True, True, 0)
        self.create_context_menu()

    def on_focus(self):
        logging.debug("novel widget focus")
        self.search_entry.grab_focus()

    def _setup_filter(self):
        self.filter = self.list_store.filter_new()
        self.filter.set_visible_func(self._novel_name_filter_func)
        self.current_filter_novel_name = None
        return self.filter

    def _novel_name_filter_func(self, model, iter, data):
        if self.current_filter_novel_name == None:
            return True
        return self.current_filter_novel_name in model[iter][1]

    def on_search_changed(self, search_entry):
        logging.debug("search changed: %s" %search_entry.get_text())
        self.current_filter_novel_name = search_entry.get_text()
        self.filter.refilter()

    def _is_novel(self, filename):
        return os.path.isfile(os.path.join("./json", filename)) and len(os.path.splitext(filename)[1]) == 0
        
    def on_item_activated(self, icon_view, path):
        model = icon_view.get_model()
        iter = model.get_iter(path)
        if iter != None:
            args = ['./NovelReader4', model[iter][1]]
            args.extend(sys.argv[1:])
            subprocess.Popen(args) #直接启动子进程，不捕获输出
            
    def on_view_novel(self, widget):
        ok, path, cell = self.icon_view.get_cursor() # 当前光标所在的path路径
        model = self.icon_view.get_model()
        if ok and path != None:
            iter = model.get_iter(path)
            if iter != None:
                f = os.popen("./NovelReader4 %s %s" %(model[iter][1], ' '.join(sys.argv[1:]))) #这儿参数也传过去
                f.close()
        
    def on_remove_novel(self, widget):
        ok, path, cell = self.icon_view.get_cursor() # 当前光标所在的path路径
        if ok and path != None:
            iter = self.filter.get_iter(path)
            if iter != None:
                novel_name = self.filter[iter][1]
                child_iter = self.filter.convert_iter_to_child_iter(iter)
                assert child_iter != None
                self.list_store.remove(child_iter) #移除图标
                self._remove_native_novel(novel_name)
                self.files.remove(novel_name) #移除缓存

    def _remove_native_novel(self, name):
        """移除本地上的书籍"""
        os.remove("./json/%s" %name)
        os.remove("./json/%s.img" %name)

    def on_update_novel(self, widget):
        logging.debug("update novel...")
        ok, path, cell = self.icon_view.get_cursor() # 当前光标所在的path路径
        if ok and path != None:
            model = self.icon_view.get_model()
            iter = model.get_iter(path)
            logging.debug("update novel '%s' from NovelsWidget" %model[iter][1])
            name = model[iter][1]
            with open('./json/%s' %name, 'rt', encoding='utf-8') as f:
                novel = json.load(f)
                novel = {'name':novel['Name'], 'author':novel['Author'], 'op':'更新'}
                self.mgr.download_or_update(novel)


    def remove_novel(self, novel):
        logging.debug("remove novel '%s'" %novel)
        iter = self._find_iter_by_name(novel)
        if iter != None:
            self.list_store.remove(iter)
            self._remove_native_novel(novel)
            self.files.remove(novel)

    def _find_iter_by_name(self, novel):
        iter = self.list_store.get_iter_first()
        while iter != None:
            if self.list_store[iter][1] == novel:
                return iter
            iter = self.list_store.iter_next(iter)
        return None
                
    def on_button_release(self, widget, event):
        if event.type == Gdk.EventType.BUTTON_RELEASE and event.button == 3:
            logging.debug("event.x=%d,event.y=%d" %(event.x, event.y))
            mouse_path = self.icon_view.get_path_at_pos(event.x, event.y) #鼠标所在项的path所在路径
            ok, path, cell = self.icon_view.get_cursor() # 当前光标所在的path路径
            logging.info("OK: %s" %ok)

            # 表明当前有项被选中
            if path != None and mouse_path != None:#这儿相当于是只要当光标在小图标上，然后点击鼠标右键才会弹出菜单
                if mouse_path == path:#表示光标在选中的项上面
                    iter = self.list_store.get_iter(path)
                    self.context_menu.popup(None, None, None, None, event.button, event.time)
            elif mouse_path == None and path != None:
                self.icon_view.unselect_path(path)

    # 来创建弹出菜单
    def create_context_menu(self):
        #首先创建ActionGroup
        action_group = Gtk.ActionGroup()

        #创建action,并且将至添加到该组中
        view_action = Gtk.Action("view_novel", "阅读", None, None)
        view_action.connect("activate", self.on_view_novel)
        action_group.add_action(view_action)

        remove_action = Gtk.Action("remove_novel", "移除", None, None)
        remove_action.connect("activate", self.on_remove_novel)
        action_group.add_action(remove_action)

        update_action = Gtk.Action("update_novel", "更新", None, None)
        update_action.connect("activate", self.on_update_novel)
        action_group.add_action(update_action)

        #创建uimanager,并且讲上面的组和他关联起来，实际上关联的是他内部从xml中加载的各个菜单的action
        uimanager = Gtk.UIManager()
        uimanager.add_ui_from_string(NovelsWidget.CONTEXT_MENU)
        uimanager.insert_action_group(action_group)

        #创建弹出菜单（其实是uimager内部获取，他已经创建好了)
        self.context_menu = uimanager.get_widget("/PopupMenu")

    def add_novel(self, name):
        if not name in self.files:
            pixbuf = Pixbuf.new_from_file_at_scale("./json/%s.img" %name, NovelsWidget.ICON_WIDTH, 
                    NovelsWidget.ICON_HEIGHT, True)
            self.list_store.append([pixbuf, name])
            self.files.append(name)

    def novel_exists(self, name):
        logging.info("novel exist......")
        logging.info("files: %s" %self.files)
        logging.info("test novel: %s" %name)
        logging.info("exist: %s" %(name in self.files))
        return (name in self.files)

class DownloadRecord:
    def __init__(self, name, author, value, op, now):
        self.name = name
        self.author = author
        self.value = value
        self.op = op
        self.now = now

class DownloadWidget(Gtk.ScrolledWindow):
    """download widget represent download page
    """

    CONTEXT_MENU = """
        <ui>
            <popup name="PopupMenu">
                <menuitem action="remove_selected_item"/>
                <menuitem action="remove_selected_item_and_file"/>
            </popup>
        </ui>
    """

    TEXT = 0
    PROGRESS = 1
    DOWNLOAD_LOG_FILE = './.download_records'

    def __init__(self, model, mgr):
        logging.info("build download widget...")
        Gtk.ScrolledWindow.__init__(self)
        self.mgr = mgr

        self.model = model
        self._setup_filter()

        self.content_treeview = Gtk.TreeView.new_with_model(self.filter) #to use filter, must set filter as the model of treeview
        self.selection = self.content_treeview.get_selection()
        self._setup_columns()

        self.add(self.content_treeview)
        self.set_policy(Gtk.PolicyType.NEVER, Gtk.PolicyType.AUTOMATIC)

        self.work_queue = WorkQueue()
        self.timeout_id = GObject.timeout_add(50, self.on_timeout, None) #一直启动定时器

        self.download_records = dict()
        self._load_download_records()

        self.connect("destroy", self.on_quit)
        self._create_context_menu()
        
        self.content_treeview.connect("button-release-event", self._on_button_release)

    def on_focus(self):
        logging.debug("download widget focus")
        pass

    # 来创建弹出菜单
    def _create_context_menu(self):
        #首先创建ActionGroup
        action_group = Gtk.ActionGroup()

        #创建action,并且将至添加到该组中
        action = Gtk.Action("remove_selected_item", "移除选中项", None, None)
        action.connect("activate", self._on_remove_selected_item)
        action_group.add_action(action)

        action = Gtk.Action("remove_selected_item_and_file", "移除选中项和下载的小说", None, None)
        action.connect("activate", self._on_remove_selected_item_and_file)
        action_group.add_action(action)

        #创建uimanager,并且讲上面的组和他关联起来，实际上关联的是他内部从xml中加载的各个菜单的action
        uimanager = Gtk.UIManager()
        uimanager.add_ui_from_string(DownloadWidget.CONTEXT_MENU)
        uimanager.insert_action_group(action_group)

        #创建弹出菜单（其实是uimager内部获取，他已经创建好了)
        self.context_menu = uimanager.get_widget("/PopupMenu")
        logging.debug("Context Menu: %s" %self.context_menu)

    # 相应右键，这儿主要是弹出菜单
    def _on_button_release(self, widget, event):
        # event.button == 3表示右键
        if event.type == Gdk.EventType.BUTTON_RELEASE and event.button == 3:
            logging.debug("event.x=%d,event.y=%d" %(event.x, event.y))
            result = self.content_treeview.get_path_at_pos(event.x, event.y) #鼠标所在项的path所在路径
            if result == None:
                return
            
            # 下面需要注意，因为使用filter来作为model的，所以应该使用filter来获取数据
            mouse_path, *_ = result
            mouse_iter = self.filter.get_iter(mouse_path)
            logging.debug("mouse novel:%s" %self.filter[mouse_iter][0])
            
            _, selected_iter = self.selection.get_selected()
            if selected_iter == None: #没有项被选中, 肯定不弹出菜单
                return

            selected_path = self.filter.get_path(selected_iter)
            logging.debug("selected novel:%s" %self.filter[selected_iter][0])

            logging.debug("mouse_iter:%s,selected_iter:%s" %(mouse_iter, selected_iter))
            logging.debug("mouse_path:%s,selected_path:%s" %(mouse_path, selected_path))

            # 下面使用迭代器来进行比较未必是相等的，但是使用path来进行比较一定是相等的
            if mouse_path == selected_path:
                self.context_menu.popup(None, None, None, None, event.button, event.time)

    def _on_remove_selected_item(self, widget):
        logging.info("remove selected item")
        _, selected_iter = self.selection.get_selected()
        if selected_iter != None:
            logging.info("remove novel '%s' item" %self.filter[selected_iter][0])
            path = self.filter.get_path(selected_iter)

            # TreeModelFilter 有个概念，他内部包装的model就做孩子
            #
            child_iter = self.filter.convert_iter_to_child_iter(selected_iter)
            logging.debug("model novel:%s" %self.model[child_iter][0])
            del self.download_records[self.model[child_iter][0]]
            self.model.remove(child_iter)

    def get_selected_novel(self):
        _, selected_iter = self.selection.get_selected()
        if selected_iter != None:
            return self.filter[selected_iter][0]
        return None 


    def _on_remove_selected_item_and_file(self, widget):
        novel = self.get_selected_novel()

        if novel == None:
            return

        self._on_remove_selected_item(widget)
        logging.debug("remove selected novel '%s'" %novel)
        self.mgr.remove_novel(novel)

    def _setup_columns(self):
        """setup treeview columns
        """
        logging.info("setup columns of download widget")
        for i, title, kind in [(0, "小说名称", DownloadWidget.TEXT), (1, "作者名称", DownloadWidget.TEXT),
                (2, "下载进度", DownloadWidget.PROGRESS), (3, "操作类型", DownloadWidget.TEXT),
                (4, "操作时间", DownloadWidget.TEXT)]:
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

    def download_or_update(self, novel):
        logging.info("download or update novel '%s'" %novel['name'])
        if novel['op'] == '更新':
            self.remove_download_record(novel['name'])

        now = time.strftime('%Y-%m-%d %H:%M:%S')
        child_iter = self.model.append([novel['name'], novel['author'], 0, novel['op'], now])

        logging.debug("download_or_update: child iter: %s, novel: %s" %(child_iter, self.model[child_iter][0]))
        #将model的iter转换为过滤器的iter
        ok, iter = self.filter.convert_child_iter_to_iter(child_iter)
        assert ok
        logging.debug("download_or_update: filter iter: %s, novel: %s" %(iter, self.filter[iter][0]))

        self.selection.select_iter(iter)

        path = self.filter.get_path(iter)
        self.content_treeview.scroll_to_cell(path, None, False, 0, 0)

        #这儿有个有趣的问题是：
        #对于filter, 他的本质上是model的包装类，
        #但是对于filter，当你在model尾部添加新的元素的时候，filter中上次保存拍的迭代器(iter)就会失效
        #但是对于model则不会失效
        _thread.start_new_thread(self.work, (child_iter, 2, novel, now))

    def work(self, iter, index, novel, now):
        """进行下载
        """
        with self._do_download_or_update(novel) as f:
            while True:
                line = f.readline() #阻塞操作
                value = self.extract_value(line)
                self.work_queue.push(Work(iter, index, value))
                if len(line) == 0:
                    break
                # 我可以在这里面抓进度，项支持断点下载, 但是关键要保证novel_download底层支持

        self.on_download_done(novel['name'])
        logging.info("Download %s done" %novel["name"])
        self.download_records[novel['name']] = DownloadRecord(novel['name'], novel['author'],
                100, novel['op'], now)

    def extract_value(self, line):
        try:
            index = line.index(">")
            return int(float(line[2:index]))
        except ValueError:
            return 100

    def _do_download_or_update(self, novel):
        if novel['op'] == "更新": #存在则进行更新
            logging.info("Update novel %s" %novel['name'])
            return os.popen("./novel_update %s" %novel['name'])
        elif novel['op'] == "下载": #否则才是下载
            logging.info("Download novel %s" %novel['name'])
            return os.popen("./novel_download_by_url %s" %novel["url"])

    def on_timeout(self, user_data):
        """到了时间，有任务就做活，没有任务就什么都不干"""
        while not self.work_queue.is_empty():
            work = self.work_queue.pop()
            self.model[work.iter][work.index] = work.value
        return True

    def _load_download_records(self):
        if not os.path.isfile(DownloadWidget.DOWNLOAD_LOG_FILE):
            return
        logging.info("load download records from '%s'" %DownloadWidget.DOWNLOAD_LOG_FILE)
        f = open(DownloadWidget.DOWNLOAD_LOG_FILE, "rt")
        while True:
            line = f.readline()
            logging.debug("line: %s" %line)
            if len(line) == 0:
                break
            line = line.strip()
            record = line.split('|')
            logging.debug("record: %s" %record)
            self.download_records[record[0]] = DownloadRecord(record[0], record[1], int(record[2]), record[3], record[4])
            logging.debug("The count of download records is %d" %len(self.download_records))
        f.close()

        # 将上次的操作写到liststore中
        for key in self.download_records:
            value = self.download_records[key]
            self.model.append([value.name, value.author, value.value, value.op, value.now])


    def _save_download_records(self):
        f = open(DownloadWidget.DOWNLOAD_LOG_FILE, "w")
        for key in self.download_records:
            value = self.download_records[key]
            f.write('%s|%s|%d|%s|%s\n' %(value.name, value.author, value.value, value.op, value.now))

    def on_quit(self, widget):
        logging.info("save download records to '%s'" %DownloadWidget.DOWNLOAD_LOG_FILE)
        self._save_download_records()

    def remove_download_record(self, name):
        """移除指定名称的record"""
        iter = self._find_novel_name(name)
        if iter != None:
            self.model.remove(iter)
            del self.download_records[name]

    def _find_novel_name(self, name):
        iter = self.model.get_iter_first()
        while iter != None:
            if self.model[iter][0] == name:
                return iter
            iter = self.model.iter_next(iter)
        return None

    def on_download_done(self, name):
        self.mgr.on_download_done(name)

class SearchWidget:

    def __init__(self, builder, mgr):
        logging.info("from build extract search widget's controls...")
        self.mgr = mgr

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

        self.download_or_update = builder.get_object("download_or_update")
        self._assert_control(self.download_or_update, "download_or_update")
        self.download_or_update.connect("clicked", self.on_download_or_update)

    def on_focus(self):
        logging.debug("search widget focus")
        self.search_content.grab_focus()

    def _assert_control(self, control, control_id):
        """assert control, ouput error msg
        """
        assert control, "can't get object '%s' from builder" %control_id

    def _on_search(self, button):
        """do search operation
        """
        logging.info('search novel "%s"' %self.search_content.get_text())
        result = subprocess.check_output([SEARCH_EXECUTED_FILE, self.search_content.get_text()])
        result = result.decode('utf-8')
        result = result.strip()

        logging.info('search result: "%s"', result)
        if result == 'None': #表明没有novel被找到
            dialog = Gtk.MessageDialog(self.mgr.window, 0, Gtk.MessageType.INFO, 
                    Gtk.ButtonsType.OK, "Novel '%s' can not be found!" %self.search_content.get_text())
            dialog.run()
            dialog.destroy()
            return

        
        self.search_result_container.show()
        result = result.split('|')
        logging.debug("icon: %s" %result[5])

        self.novel = dict()

        self.novel['url'] = result[0]
        self.url = result[0]
        self.novel['name'] = result[1]
        self.novel_name.set_text(result[1])
        self.novel['author'] = result[2]
        self.novel_author_name.set_text(result[2])

        desp = self._filter_description(result[3])
        self.novel['description'] = desp
        self.novel_description.set_text(desp)
        self.novel['lastchapter'] = result[4]
        self.novel_last_chapter_name.set_text(result[4])

        logging.debug("file %s exists? %s" %(result[5], os.path.isfile(result[5])))
        assert os.path.isfile(result[5]), 'icon file not exists'

        logging.debug("set novel image: '%s'" %result[5])
        pixbuf = Pixbuf.new_from_file_at_scale(result[5], 150, 
                200, True)
        self.novel_img.set_from_pixbuf(pixbuf)
        if self.mgr.novel_exists(result[1]):
            self.download_or_update.set_label("更新")
        else:
            self.download_or_update.set_label("下载")
        self.novel['op'] = self.download_or_update.get_label()

    def _filter_description(self, desp):
        lines = desp.split('\n')
        lines = ['    %s' %line.strip() for line in lines]
        return '\n'.join(lines)

    def on_download_or_update(self, button):
        self.mgr.download_or_update(self.novel)

    def show(self):
        self.widget.show()

    def hide(self):
        self.widget.hide()

class MainWindow:

    def __init__(self):
        logging.info("parse glade file '%s'..." %'./app.glade')
        builder = Gtk.Builder()
        builder.add_from_file(os.path.join(THIS_SCRIPT_DIRNAME, "app.glade"))
        builder.connect_signals(self)

        self._init_controls(builder)
        self._setup_menu()
        self._set_style()
    
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

        self.download_widget = DownloadWidget(self.download_liststore, self)
        self.content_container.pack_start(self.download_widget, True, True, 0)

        self.novels_widget = NovelsWidget(self)
        self.content_container.pack_start(self.novels_widget, True, True, 0)

        builder.get_object('content_container').set_name('content_container') #设置name可以再css中使用#进行引用

    def _setup_menu(self):
        """setup tree view menu
        """
        # add search menu item
        self.menu_treestore.append(None, ['搜索'])

        # add download menu item
        download_parent = self.menu_treestore.append(None, ['下载'])
        for item in ('正在下载', '已经完成', '更新操作', '下载操作'):
            self.menu_treestore.append(download_parent, [item])

        # add novel menu item
        self.menu_treestore.append(None, ['小说'])


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
                self.novels_widget.hide()
                self.search_widget.on_focus()
            elif name in ('下载', '正在下载', '已经完成', '更新操作', '下载操作'):
                self.download_widget.show()
                self.search_widget.hide()
                self.novels_widget.hide()
                self.download_widget.switch(name)
                self.download_widget.on_focus()
            elif name == "小说":
                self.novels_widget.show()
                self.download_widget.hide()
                self.search_widget.hide()
                self.novels_widget.on_focus()

    def novel_exists(self, name):
        return self.novels_widget.novel_exists(name)

    def download_or_update(self, novel):
        path = Gtk.TreePath(1)
        iter = self.menu_treestore.get_iter(path)
        self.menu_selection.select_iter(iter)
        self.download_widget.download_or_update(novel)

    def on_download_done(self, name):
        self.novels_widget.add_novel(name)

    # 下面一个地方设置，搜索，所有自检都会被控制
    def _set_style(self):
        cssProvider = Gtk.CssProvider.new()
        if cssProvider.load_from_path("./app.css"):
            self._apply_style(self.window, cssProvider)

    def _apply_style(self, widget, provider):
        context = widget.get_style_context()
        context.add_provider(provider, Gtk.STYLE_PROVIDER_PRIORITY_USER)
        if isinstance(widget, Gtk.Container) :
            widget.forall(self._apply_style, provider)

    def remove_novel(self, novel_name):
        self.novels_widget.remove_novel(novel_name)

def main():
    config.config()

    win = MainWindow()
    win.show()

    Gtk.main()
        
if __name__ == '__main__':
    main()
