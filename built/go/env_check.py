#coding=utf-8
import os
import sys

if len(sys.argv) < 2:
    print("not include version info")
    sys.exit(2)

expect_version = 'go'+sys.argv[1]
print("expect_version: " + expect_version)

command = 'go version'
r = os.popen(command)
info = r.readlines()
version = info[0].strip('\r\n')
go_version = version.split(' ')[2]

if not (go_version >= expect_version):
    print("real go version: " + go_version)
    sys.exit(1)