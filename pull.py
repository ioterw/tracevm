#!/usr/bin/python3
import os, subprocess

def popen(*args, **kwargs):
    p = subprocess.Popen(*args, **kwargs)
    res = p.communicate()
    if p.returncode != 0:
        exit(1)
    return res

def main():
    os.chdir(os.path.dirname(os.path.abspath(__file__)) + '/go-ethereum')
    popen(['git', 'pull'])

if __name__ == '__main__':
    main()
