# -*- coding: utf-8 -*-

from __future__ import absolute_import

import socks


def send_to():
    s = socks.socksocket()
    s.set_proxy(socks.SOCKS5, "localhost", 1080)
    s.connect(("www.baidu.com", 80))
    s.sendall(("GET / HTTP/1.1"))
    print s.recv(4096)
