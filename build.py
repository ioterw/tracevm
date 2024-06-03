#!/usr/bin/python3
import os, subprocess, shutil, glob

def popen(*args, **kwargs):
    p = subprocess.Popen(*args, **kwargs)
    res = p.communicate()
    if p.returncode != 0:
        exit(1)
    return res

def main():
    target = 'geth'
    
    if target == 'geth':
        os.chdir(os.path.dirname(os.path.abspath(__file__)) + '/go-ethereum')

        shutil.rmtree('eth/tracers/live/dep_tracer',  ignore_errors=True)
        shutil.rmtree('eth/tracers/live/dep_geth.go', ignore_errors=True)

        shutil.copytree('../tracer/dep_tracer', 'eth/tracers/live/dep_tracer')
        shutil.copy('../tracer/dep_geth.go', 'eth/tracers/live/dep_geth.go')

        popen(['go', 'get', 'github.com/basho/riak-go-client'])
        popen(['make', 'geth'])
        shutil.copy('build/bin/geth', '..')
    else:
        print('Unknown target:', target)
        exit(1)

    os.chdir('..')
    shutil.rmtree('build', ignore_errors=True)
    os.mkdir('build')
    shutil.copy('run.py', 'build')
    shutil.copy('geth', 'build')
    shutil.copy('conf_examples/default.json', 'conf.json')
    shutil.copy('conf.json', 'build')

if __name__ == '__main__':
    main()
