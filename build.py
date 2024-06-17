#!/usr/bin/python3
import os, subprocess, shutil, glob, sys, re, stat

def popen(*args, **kwargs):
    p = subprocess.Popen(*args, **kwargs)
    res = p.communicate()
    if p.returncode != 0:
        exit(1)
    return res

def mkdir(path):
    try:
        os.mkdir(path)
    except FileExistsError:
        pass

def rmfile(path):
    try:
        os.remove(path)
    except FileNotFoundError:
        pass

def reset_file(path):
    popen(['git', 'checkout', 'HEAD', '--', path])

def patch_file(path, pattern, replacement):
    with open(path, 'r') as f:
        data0 = f.read()
    data1 = re.sub(pattern, replacement, data0)
    if data0 == data1:
        print('Patch', path, 'failed')
        exit(1)
    with open(path, 'w') as f:
        f.write(data1)

def reset_and_patch_files(path, *args):
    reset_file(path)
    for pattern, replacement in args:
        patch_file(path, pattern, replacement)

def build_geth(root):
    os.chdir(root + '/go-ethereum')

    shutil.rmtree('eth/tracers/live/dep_tracer',  ignore_errors=True)
    rmfile('eth/tracers/live/geth_dep.go')

    mkdir('eth/tracers/live/dep_tracer')
    for path in glob.iglob('../tracer/dep_tracer/*.go'):
        shutil.copy(path, 'eth/tracers/live/dep_tracer/')
    shutil.copy('../tracer/extra/geth_dep.go', 'eth/tracers/live/geth_dep.go')

    popen(['go', 'get', 'github.com/basho/riak-go-client'])
    popen(['make', 'geth'])

    os.chdir('..')
    mkdir('build')

    shutil.copy('go-ethereum/build/bin/geth', 'build/geth')
    shutil.copy('conf_examples/default.json', 'build/conf.json')
    shutil.copy('tracer/extra/geth_run.py', 'build/run.py')

    st = os.stat('build/run.py')
    os.chmod('build/run.py', st.st_mode | stat.S_IEXEC)

def build_lib(root):
    os.chdir(root + '/tracer')

    mkdir('../build')

    rmfile('../build/libdep.a')
    rmfile('../build/libdep.h')

    env = os.environ.copy()
    env['CGO_ENABLED'] = '1'
    popen(['go', 'build', '-buildmode=c-archive', 'libdep.go'], env=env)
    shutil.move('libdep.a', '../build')
    shutil.move('libdep.h', '../build')

def build_foundry(root):
    build_lib(root)

    os.chdir(root + '/foundry/revm-inspectors')
    reset_and_patch_files(
        'Cargo.toml',
        (
            r'\n\[dependencies\]\n',
            (
                '\n'
                '[dependencies]\n'
                'tracevm = { path = "../tracevm" }\n'
            ),
        ),
    )

    os.chdir(root + '/foundry/foundry')
    reset_and_patch_files(
        'Cargo.toml',
        (
            r'rust\-version \= "1\.76"',
            'rust-version = "1.77.0"',
        ), (
            r'revm\-inspectors\s*\=\s*\{\s*git\s*\=\s*"https\://github\.com/paradigmxyz/revm\-inspectors"\s*\,\s*rev\s*\=\s*"5cf339c"\s*\,\s*features\s*\=\s*\[\s*"serde"\s*,\s*\]\s*}',
            'revm-inspectors = { path = "../revm-inspectors", features = ["serde"] }',
        ), (
            r'\n\[workspace.dependencies\]\n',
            (
                '\n'
                '[workspace.dependencies]\n'
                'tracevm = { path = "../tracevm" }\n'
            ),
        )
    )
    reset_and_patch_files(
        'crates/cast/Cargo.toml',
        (
            r'\n\[dependencies\]\n',
            (
                '\n'
                '[dependencies]\n'
                'tracevm.workspace = true\n'
            ),
        ),
    )
    reset_and_patch_files(
        'crates/cast/bin/cmd/run.rs',
        (
            r'use alloy_primitives::U256;\n',
            (
                'use tracevm;\n'
                'use alloy_primitives::U256;\n'
            ),
        ), (
            r'\n        let result \= \{\n',
            (
                '\n'
                '        tracevm::activate(tx.hash);\n'
                '        let result = {\n'
            ),
        ),
    )
    reset_and_patch_files(
        'crates/evm/evm/Cargo.toml',
        (
            r'\n\[dependencies\]\n',
            (
                '\n'
                '[dependencies]\n'
                'tracevm.workspace = true\n'
            ),
        ),
    )
    reset_and_patch_files(
        'crates/evm/evm/src/inspectors/debugger.rs',
        (
            r'use alloy_primitives::Address;\n',
            (
                'use tracevm;\n'
                'use alloy_primitives::Address;\n'
            ),
        ), (
            r'\npub struct Debugger \{\n',
            (
                '\n'
                'pub struct Debugger {\n'
                '    dep_data: tracevm::DepData<{tracevm::DepDataType::Debug as u8}>,\n'
            ),
        ), (
            r'\n    fn step\(\s*&mut self\,\s*interp\: &mut Interpreter\,\s*ecx\: &mut EvmContext<DB>\,?\s*\)\s*\{\n',
            (
                '\n'
                '    fn step_end(&mut self, interp: &mut Interpreter, ecx: &mut EvmContext<DB>) {\n'
                '        tracevm::dep_step_end(&mut self.dep_data, interp, ecx);\n'
                '    }\n'
                '\n'
                '    fn step(&mut self, interp: &mut Interpreter, ecx: &mut EvmContext<DB>) {\n'
                '        tracevm::dep_step(&mut self.dep_data, interp, ecx);\n'
            ),
        ), (
            r'\n    fn call\(\s*&mut self\,\s*ecx\: &mut EvmContext<DB>\,\s*inputs\: &mut CallInputs\,?\s*\)\s*\->\s*Option<CallOutcome>\s*\{\n',
            (
                '\n'
                '    fn call(&mut self, ecx: &mut EvmContext<DB>, inputs: &mut CallInputs) -> Option<CallOutcome> {\n'
                '        tracevm::dep_call(&mut self.dep_data, ecx, inputs);\n'
            ),
        ), (
            r'\n    fn call_end\(\s*&mut self\,\s*_context\: &mut EvmContext<DB>\,\s*_inputs\: &CallInputs\,\s*outcome\: CallOutcome\,\s*\)\s*\->\s*CallOutcome\s*\{\n',
            (
                '\n'
                '    fn call_end( &mut self, _context: &mut EvmContext<DB>, _inputs: &CallInputs, outcome: CallOutcome) -> CallOutcome {\n'
                '        tracevm::dep_call_end(&mut self.dep_data, _context, _inputs, &outcome);\n'
            ),
        ), (
            r'\n    fn create\(\s*&mut self\,\s*ecx\: &mut EvmContext<DB>\,\s*inputs\: &mut CreateInputs\,?\s*\)\s*\->\s*Option<CreateOutcome>\s*\{\n',
            (
                '\n'
                '    fn create( &mut self, ecx: &mut EvmContext<DB>, inputs: &mut CreateInputs) -> Option<CreateOutcome> {\n'
                '        tracevm::dep_create(&mut self.dep_data, ecx, inputs);\n'
            ),
        ), (
            r'\n    fn create_end\(\s*&mut self\,\s*_context\: &mut EvmContext<DB>\,\s*_inputs\: &CreateInputs\,\s*outcome\: CreateOutcome\,?\s*\)\s*\->\s*CreateOutcome\s*\{\n',
            (
                '\n'
                '    fn create_end( &mut self, _context: &mut EvmContext<DB>, _inputs: &CreateInputs, outcome: CreateOutcome) -> CreateOutcome {\n'
                '        tracevm::dep_create_end(&mut self.dep_data, _context, _inputs, &outcome);\n'
            ),
        ),
    )

    env = os.environ.copy()
    env['DEP_PATH'] = f'{root}/build'
    env['RUSTFLAGS'] = f'-Clink-args=-Wl,-rpath,{env["DEP_PATH"]}'
    os.chdir('crates/cast')
    popen(['cargo', 'build'], env=env)
    os.chdir('../..')

    shutil.copy('target/debug/cast', '../../build/tracevm-cast')

def main(target):
    root = os.path.dirname(os.path.abspath(__file__))
    target = target.strip('/')
    if target == 'all':
        build_geth(root)
        build_foundry(root)
    elif target in ['geth', 'go-ethereum']:
        build_geth(root)
    elif target == 'lib':
        build_lib(root)
    elif target == 'foundry':
        build_foundry(root)
    else:
        print('Unknown target:', target)
        print('Targets: all, geth, lib, foundry')
        exit(1)

if __name__ == '__main__':
    if len(sys.argv) > 1:
        main(sys.argv[1])
    else:
        main('all')
