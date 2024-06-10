#!/usr/bin/python3
import os, subprocess, shutil, glob, sys, re

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

def build_geth(root):
    os.chdir(root + '/go-ethereum')

    shutil.rmtree('eth/tracers/live/dep_tracer',  ignore_errors=True)
    rmfile('eth/tracers/live/dep_geth.go')

    mkdir('eth/tracers/live/dep_tracer')
    for path in glob.iglob('../tracer/dep_tracer/*.go'):
        shutil.copy(path, 'eth/tracers/live/dep_tracer/')
    shutil.copy('../tracer/extra/dep_geth.go', 'eth/tracers/live/dep_geth.go')

    popen(['go', 'get', 'github.com/basho/riak-go-client'])
    popen(['make', 'geth'])
    shutil.copy('build/bin/geth', '..')

    os.chdir('..')
    mkdir('build')
    shutil.copy('run.py', 'build')
    shutil.copy('geth', 'build')
    shutil.copy('conf_examples/default.json', 'conf.json')
    shutil.copy('conf.json', 'build')

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

    os.chdir(root + '/foundry')

    shutil.copy('../tracer/extra/build_foundry.rs', 'crates/evm/evm/build.rs')
    mkdir('crates/evm/evm/src/inspectors/debugger')
    shutil.copy('../tracer/extra/mod_foundry.rs', 'crates/evm/evm/src/inspectors/debugger/dep_tracer.rs')

    reset_file('Cargo.toml')
    patch_file(
        'Cargo.toml',
        r'rust\-version \= "1\.76"',
        'rust-version = "1.77.0"',
    )

    reset_file('crates/cast/bin/cmd/run.rs')
    patch_file(
        'crates/cast/bin/cmd/run.rs',
        r'use alloy_primitives::U256;\n',
        (
            'use foundry_evm::inspectors::debugger::dep_tracer;\n'
            'use alloy_primitives::U256;\n'
        ),
    )
    patch_file(
        'crates/cast/bin/cmd/run.rs',
        r'\n        let result \= \{\n',
        (
            '\n'
            '        dep_tracer::activate(tx.hash);\n'
            '        let result = {\n'
        ),
    )



    reset_file('crates/evm/evm/src/inspectors/mod.rs')
    patch_file(
        'crates/evm/evm/src/inspectors/mod.rs',
        r'\nmod debugger;\n',
        (
            '\n'
            'pub mod debugger;\n'
        ),
    )

    reset_file('crates/evm/evm/src/inspectors/debugger.rs')
    patch_file(
        'crates/evm/evm/src/inspectors/debugger.rs',
        r'\npub struct Debugger \{\n',
        (
            '\n'
            'pub struct Debugger {\n'
            '    dep_data: dep_tracer::DepData,\n'
        ),
    )
    patch_file(
        'crates/evm/evm/src/inspectors/debugger.rs',
        r'\nuse revm_inspectors\:\:tracing\:\:types\:\:CallKind;\n',
        (
            '\n'
            'use revm_inspectors::tracing::types::CallKind;\n'
            'pub mod dep_tracer;\n'
        ),
    )
    patch_file(
        'crates/evm/evm/src/inspectors/debugger.rs',
        r'\n    fn step\(\s*&mut self\,\s*interp\: &mut Interpreter\,\s*ecx\: &mut EvmContext<DB>\,?\s*\)\s*\{\n',
        (
            '\n'
            '    fn step_end(&mut self, interp: &mut Interpreter, ecx: &mut EvmContext<DB>) {\n'
            '        dep_tracer::dep_step_end(&mut self.dep_data, interp, ecx);\n'
            '    }\n'
            '\n'
            '    fn step(&mut self, interp: &mut Interpreter, ecx: &mut EvmContext<DB>) {\n'
            '        dep_tracer::dep_step(&mut self.dep_data, interp, ecx);\n'
        ),
    )
    patch_file(
        'crates/evm/evm/src/inspectors/debugger.rs',
        r'\n    fn call\(\s*&mut self\,\s*ecx\: &mut EvmContext<DB>\,\s*inputs\: &mut CallInputs\,?\s*\)\s*\->\s*Option<CallOutcome>\s*\{\n',
        (
            '\n'
            '    fn call(&mut self, ecx: &mut EvmContext<DB>, inputs: &mut CallInputs) -> Option<CallOutcome> {\n'
            '        dep_tracer::dep_call(&mut self.dep_data, ecx, inputs);\n'
        ),
    )
    patch_file(
        'crates/evm/evm/src/inspectors/debugger.rs',
        r'\n    fn call_end\(\s*&mut self\,\s*_context\: &mut EvmContext<DB>\,\s*_inputs\: &CallInputs\,\s*outcome\: CallOutcome\,\s*\)\s*\->\s*CallOutcome\s*\{\n',
        (
            '\n'
            '    fn call_end( &mut self, _context: &mut EvmContext<DB>, _inputs: &CallInputs, outcome: CallOutcome) -> CallOutcome {\n'
            '        dep_tracer::dep_call_end(&mut self.dep_data, _context, _inputs, &outcome);\n'
        ),
    )
    patch_file(
        'crates/evm/evm/src/inspectors/debugger.rs',
        r'\n    fn create\(\s*&mut self\,\s*ecx\: &mut EvmContext<DB>\,\s*inputs\: &mut CreateInputs\,?\s*\)\s*\->\s*Option<CreateOutcome>\s*\{\n',
        (
            '\n'
            '    fn create( &mut self, ecx: &mut EvmContext<DB>, inputs: &mut CreateInputs) -> Option<CreateOutcome> {\n'
            '        dep_tracer::dep_create(&mut self.dep_data, ecx, inputs);\n'
        ),
    )
    patch_file(
        'crates/evm/evm/src/inspectors/debugger.rs',
        r'\n    fn create_end\(\s*&mut self\,\s*_context\: &mut EvmContext<DB>\,\s*_inputs\: &CreateInputs\,\s*outcome\: CreateOutcome\,?\s*\)\s*\->\s*CreateOutcome\s*\{\n',
        (
            '\n'
            '    fn create_end( &mut self, _context: &mut EvmContext<DB>, _inputs: &CreateInputs, outcome: CreateOutcome) -> CreateOutcome {\n'
            '        dep_tracer::dep_create_end(&mut self.dep_data, _context, _inputs, &outcome);\n'
        ),
    )

    env = os.environ.copy()
    env['DEP_PATH'] = f'{root}/build'
    env['RUSTFLAGS'] = f'-Clink-args=-Wl,-rpath,{env["DEP_PATH"]}'
    os.chdir('crates/cast')
    popen(['cargo', 'build'], env=env)
    os.chdir('../..')

    shutil.copy('target/debug/cast', '../build/tracevm-cast')

def main(target):
    root = os.path.dirname(os.path.abspath(__file__))
    if target == 'geth':
        build_geth(root)
    elif target == 'lib':
        build_lib(root)
    elif target == 'foundry':
        build_foundry(root)
    else:
        print('Unknown target:', target)
        print('Targets: geth, lib, foundry')
        exit(1)

if __name__ == '__main__':
    main(sys.argv[1])
