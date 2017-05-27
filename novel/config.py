#!/usr/bin/env python3
import logging, getopt, sys

def config(*args, **kwargs):
    filename = None
    if 'filename' in kwargs:
        filename = kwargs['filename']

    if len(args) == 0:
        cmd_args = sys.argv[1:]
    else:
        cmd_args = args[0]

    config_logging(cmd_args, filename)

def config_logging(cmd_args, filename):
    str_level = parse_cmdline(cmd_args)
    log_level = getattr(logging, str_level.upper(), None)
    logging.basicConfig(filename=filename, level=log_level, format="%(asctime)s [%(levelname)s]: %(message)s",
            datefmt="%Y年%m月%d日 %H:%M:%S")

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
