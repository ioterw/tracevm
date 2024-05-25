#!/usr/bin/python3
import os, subprocess, shutil, glob

def popen(*args, **kwargs):
    p = subprocess.Popen(*args, **kwargs)
    res = p.communicate()
    if p.returncode != 0:
        exit(1)
    return res

def main():
    os.chdir(os.path.dirname(os.path.abspath(__file__)) + '/go-ethereum')
    for path in glob.glob('../tracer/*'):
        file_name = path.rsplit('/', 1)[-1]
        file_path = 'eth/tracers/live/' + file_name
        shutil.rmtree(file_path, ignore_errors=True)
        if os.path.isfile(path):
            shutil.copy(path, file_path)
        else:
            shutil.copytree(path, file_path)

    popen(['go', 'get', 'github.com/basho/riak-go-client'])
    popen(['make', 'geth'])
    shutil.copy('build/bin/geth', '..')

    os.chdir('..')
    shutil.rmtree('build', ignore_errors=True)
    os.mkdir('build')
    shutil.copy('run.py', 'build')
    shutil.copy('geth', 'build')
    shutil.copy('conf.json', 'build')
    shutil.copytree('webview', 'build/webview')

if __name__ == '__main__':
    main()
