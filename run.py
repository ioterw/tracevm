#!/usr/bin/python3
import os, subprocess, json, sys

def popen(*args, **kwargs):
    p = subprocess.Popen(*args, **kwargs)
    res = p.communicate()
    if p.returncode != 0:
        exit(1)
    return res

def handle_config(config_path):
    json_folder = os.path.abspath(config_path).rsplit('/', 1)[0]
    with open(config_path, 'r') as f:
        config = json.load(f)
    return config

def run_geth(config):
    geth_path = os.path.dirname(os.path.abspath(__file__)) + '/geth'
    popen([
        geth_path,
        '--vmtrace', 'dep', '--vmtrace.jsonconfig', json.dumps(config),
        '--dev', '--nodiscover', '--maxpeers', '0', '--mine',
        '--http', '--http.corsdomain', '*', '--http.api', 'web3,eth,debug,personal,net',
    ])

def main(config_path):
    config = handle_config(config_path)
    run_geth(config)

if __name__ == '__main__':
    main(sys.argv[1])
