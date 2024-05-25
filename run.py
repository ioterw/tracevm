#!/usr/bin/python3
import os, subprocess, json, sys, shutil, webbrowser, tempfile
import webview.app

def popen(*args, **kwargs):
    p = subprocess.Popen(*args, **kwargs)
    res = p.communicate()
    if p.returncode != 0:
        exit(1)
    return res

def handle_config(config_path):
    remove_db_path = None

    json_folder = os.path.abspath(config_path).rsplit('/', 1)[0]
    with open(config_path, 'r') as f:
        config = json.load(f)
    if config['kv']['engine'] == 'leveldb':
        if not config['kv']['root'].startswith('/'):
            config['kv']['root'] = json_folder + '/' + config['kv']['root']
        remove_db_path = config['kv']['root']
    if 'output' in config and config['output'] != '' and not config['output'].startswith('/'):
        config['output'] = json_folder + '/' + config['output']
    return config, remove_db_path

def run_server(config):
    app_path = os.path.dirname(os.path.abspath(__file__)) + '/webview/app.py'
    p = subprocess.Popen(['python3', app_path, config['output']], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    return p

def run_geth(config, remove_db_path):
    vmtrace_option = []
    geth_path = os.path.dirname(os.path.abspath(__file__)) + '/geth'
    popen([
        geth_path,
        '--vmtrace', 'dep', '--vmtrace.jsonconfig', json.dumps(config),
        '--dev', '--nodiscover', '--maxpeers', '0', '--mine',
        '--http', '--http.corsdomain', '*', '--http.api', 'web3,eth,debug,personal,net',
    ])

def main(config_path):
    config, remove_db_path = handle_config(config_path)
    try:
        p = run_server(config)
        webbrowser.open('http://127.0.0.1:4334', new=0, autoraise=True)
        run_geth(config, remove_db_path)
    finally:
        p.kill()

if __name__ == '__main__':
    main(sys.argv[1])
