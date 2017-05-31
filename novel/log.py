#!/usr/bin/env python3
import logging, getopt, sys
import config as cfg
from logging.handlers import RotatingFileHandler

FORMATTER = logging.Formatter("%(asctime)s [%(levelname)s]: %(message)s")

def config(*args):
    if len(args) == 0:
        cmd_args = sys.argv[1:]
    else:
        cmd_args = args[0]

    # 这儿设置根日志总过滤级别为全开
    logging.getLogger('').setLevel(logging.DEBUG) #总开关必须开，方便后面的handler进行再过滤

    config_logging(cmd_args)
    config_rotating_logging()

def config_rotating_logging():
    #设置本地回滚日志，即使命令行没有制定--log=debug，他内部也会自动记录日志的
    rotating_handler = RotatingFileHandler('%s/%s' %(cfg.config['log_dirname'], cfg.config['log_filename']), maxBytes=cfg.config['max_log_file_size'],
            backupCount=cfg.config["log_backup_count"])

    rotating_handler.setLevel(logging.DEBUG)
    rotating_handler.setFormatter(FORMATTER)
    logging.getLogger("").addHandler(rotating_handler)


def config_logging(cmd_args):
    str_level = parse_cmdline(cmd_args)
    log_level = getattr(logging, str_level.upper(), None)

    console_handler = logging.StreamHandler()
    console_handler.setLevel(log_level)
    console_handler.setFormatter(FORMATTER)
    logging.getLogger('').addHandler(console_handler) #logging.getLogger('')获取的是全局的logger


def parse_cmdline(cmd_args):
    try:
        opts, args = getopt.getopt(cmd_args, "l:", ["log="])
    except getopt.GetoptError:
        print("Usage: %s --log=[DEBUG|INFO|WARNING|ERROR|CRITICAL]" %sys.argv[0])
        sys.exit(1)

    for opt, arg in opts:
        if opt in ("-l", "--log"):
            return arg
    return "WARNING"
