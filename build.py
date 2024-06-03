#!/usr/bin/python3
import os, subprocess, shutil, glob, sys

def popen(*args, **kwargs):
    p = subprocess.Popen(*args, **kwargs)
    res = p.communicate()
    if p.returncode != 0:
        exit(1)
    return res

def main(target):
    if target == 'geth':
        os.chdir(os.path.dirname(os.path.abspath(__file__)) + '/go-ethereum')

        shutil.rmtree('eth/tracers/live/dep_tracer',  ignore_errors=True)
        shutil.rmtree('eth/tracers/live/dep_geth.go', ignore_errors=True)

        os.mkdir('eth/tracers/live/dep_tracer')
        for path in glob.iglob('../tracer/dep_tracer/*.go'):
            shutil.copy(path, 'eth/tracers/live/dep_tracer/')
        shutil.copy('../tracer/dep_geth.go', 'eth/tracers/live/dep_geth.go')

        popen(['go', 'get', 'github.com/basho/riak-go-client'])
        popen(['make', 'geth'])
        shutil.copy('build/bin/geth', '..')

        os.chdir('..')
        try:
            os.mkdir('build')
        except FileExistsError:
            pass
        shutil.copy('run.py', 'build')
        shutil.copy('geth', 'build')
        shutil.copy('conf_examples/default.json', 'conf.json')
        shutil.copy('conf.json', 'build')
    elif target == 'lib':
        os.chdir(os.path.dirname(os.path.abspath(__file__)) + '/tracer/dep_tracer')

        try:
            os.mkdir('../../build')
        except FileExistsError:
            pass
        
        popen(['go', 'build', '-buildmode=archive', '-o', '../../build/lib.a'])
    else:
        print('Unknown target:', target)
        print('Targets: geth, lib')
        exit(1)

if __name__ == '__main__':
    main(sys.argv[1])
