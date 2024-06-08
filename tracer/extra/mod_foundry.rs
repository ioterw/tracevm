use std::ffi::CString;
use revm::{
    EvmContext,
    interpreter::{ Interpreter, OpCode, CallInputs, CallOutcome, CreateInputs, CreateOutcome, InterpreterResult },
    primitives::{ TransactTo, SpecId },
};
use foundry_evm_core::{ backend::DatabaseExt };
use alloy_primitives::{ address, Address, Bytes, U256 };

#[repr(C)]
struct CAddress {
    data: [u8; 20],
}
fn address_to_caddress(addr: Address) -> CAddress {
    CAddress{ data: *addr.as_ref() }
}
#[repr(C)]
#[derive(Clone)]
struct CHash {
    data: [u8; 32],
}
fn deadbeef_chash() -> CHash {
    CHash {
        data: [
            0xde,0xad,0xbe,0xef,0xde,0xad,0xbe,0xef,0xde,0xad,0xbe,0xef,0xde,0xad,0xbe,0xef,
            0xde,0xad,0xbe,0xef,0xde,0xad,0xbe,0xef,0xde,0xad,0xbe,0xef,0xde,0xad,0xbe,0xef
        ],
    }
}
#[repr(C)]
struct CSizedArray {
    data: *const u8,
    size: i32,
}
fn bytes_to_csizedarray(data: &Bytes) -> CSizedArray {
    CSizedArray {
        data: data.as_ptr(),
        size: data.len() as i32,
    }
}
#[repr(C)]
struct CStack {
    data: *const CHash,
    size: i32,
}
fn stack_to_cstack(stack: &Vec<U256>) -> CStack {
    let mut cstack = vec![deadbeef_chash(); stack.len()];
    for i in 0..stack.len() {
        cstack[i] = CHash{ data: stack[i].to_be_bytes() }
    }
    CStack {
        data: cstack.as_ptr(),
        size: stack.len() as i32,
    }
}

extern "C" {
    fn RegisterGetNonce(ptr: extern "C" fn(CAddress) -> u64);
    fn RegisterGetCode(ptr: extern "C" fn(CAddress) -> CSizedArray);

    fn InitDep(cfg: *const i8);

    fn StartTransactionRecording(
        isCreate: bool, addr: CAddress, input: CSizedArray, block: u64,
        timestamp: u64, origin: CAddress, txHash: CHash,
        code: CSizedArray, isSelfdestruct6780: bool, isRandom: bool,
    );
    fn EndTransactionRecording();

    fn HandleOpcode(
        stack: CStack, memory: CSizedArray, addr: CAddress,
        pc: u64, op: u8, isInvalid: bool, hasError: bool,
    );
    fn HandleEnter(to: CAddress, input: CSizedArray);
    fn HandleExit(output: CSizedArray, hasError: bool);
    // fn HandleFault(op: u8);
}

#[derive(Clone, Debug)]
pub(crate) struct DepData {
    pub call_depth: i32,
}
impl Default for DepData {
    fn default() -> DepData {
        let cfg = CString::new(
            "{
                \"kv\": {\"engine\": \"amnesia\", \"root\": \"\"}, 
                \"logger\": {
                    \"opcodes_short\": [\"e0\", \"e1\"],
                    \"opcodes\": [],
                    \"final_slots_short\": true,
                    \"final_slots\": false,
                    \"codes_short\": false,
                    \"codes\": false,
                    \"return_data_short\": false,
                    \"return_data\": false,
                    \"logs_short\": false,
                    \"logs\": false,
                    \"sol_view\": true
                },
                \"output\": \"http://0.0.0.0:4334\",
                \"past_unknown\": true
            }"
        ).expect("CString::new failed");
        unsafe {
            InitDep(cfg.as_ptr());
            RegisterGetNonce(get_nonce);
            RegisterGetCode(get_code);
        }
        DepData {
            call_depth: 0,
        }
    }
}

static ZERO_ADDRESS: Address = address!("0000000000000000000000000000000000000000");

extern "C"
fn get_nonce(_addr: CAddress) -> u64 {
    unsafe {
        if GET_NONCE_ADDRESS == ZERO_ADDRESS {
            panic!("GET_NONCE_ADDRESS is zero")
        }
        GET_NONCE_ADDRESS = ZERO_ADDRESS;
        GET_NONCE_NONCE
    }
}
static mut GET_NONCE_ADDRESS: Address = ZERO_ADDRESS;
static mut GET_NONCE_NONCE:   u64     = 0;

extern "C"
fn get_code(_addr: CAddress) -> CSizedArray {
    unsafe {
        if GET_CODE_ADDRESS == ZERO_ADDRESS {
            panic!("GET_CODE_ADDRESS is zero")
        }
        GET_CODE_ADDRESS = ZERO_ADDRESS;
        CSizedArray{
            data: GET_CODE_DATA.as_ptr(),
            size: GET_CODE_DATA.len() as i32,
        }
    }
}
static mut GET_CODE_ADDRESS: Address = ZERO_ADDRESS;
static mut GET_CODE_DATA:    Bytes   = Bytes::new();


fn on_enter<DB:DatabaseExt>(data: &mut DepData, context: &mut EvmContext<DB>, is_create: bool, input: &Bytes) {
    let addr: Address;
    let origin = context.inner.env.tx.caller;
    if is_create {
        addr = origin.create(context.inner.env.tx.nonce.expect("nonce is missing"));
    } else {
        let TransactTo::Call(tmp_addr) = context.inner.env.tx.transact_to else { panic!("impossible create") };
        addr = tmp_addr;
    }

    if data.call_depth == 0 {
        let input = context.inner.env.tx.data.clone();
        let code: Bytes;
        if is_create {
            code = Bytes::new()
        } else {
            let (bytecode, _) = context.code(addr).unwrap();
            code = bytecode.bytecode().clone();
        }
        let block = context.inner.env.block.number;
        let timestamp = context.inner.env.block.timestamp;
        let is_selfdestruct6780 = SpecId::enabled(context.inner.journaled_state.spec, SpecId::CANCUN);
        let is_random = context.inner.env.block.prevrandao.is_some();

        unsafe {
            StartTransactionRecording(
                is_create,
                address_to_caddress(addr),
                bytes_to_csizedarray(&input),
                block.to::<u64>(),
                timestamp.to::<u64>(),
                address_to_caddress(origin),
                deadbeef_chash(),
                bytes_to_csizedarray(&code),
                is_selfdestruct6780,
                is_random,
            );
        }
    } else {
        unsafe {
            HandleEnter(
                address_to_caddress(addr),
                bytes_to_csizedarray(&input),
            )
        }
    }

    data.call_depth += 1;
}

fn on_exit<DB:DatabaseExt>(data: &mut DepData, _context: &mut EvmContext<DB>, result: &InterpreterResult) {
    data.call_depth -= 1;

    if data.call_depth == 0 {
        unsafe {
            EndTransactionRecording();
        }
    } else {
        unsafe {
            // interp.return_data_buffer
            HandleExit(
                bytes_to_csizedarray(&result.output),
                false,
            )
        }
    }
}

pub(crate) fn dep_step<DB:DatabaseExt>(_data: &mut DepData, interp: &mut Interpreter, context: &mut EvmContext<DB>) {
    let is_invalid: bool;
    if let Some(op) = OpCode::new(interp.current_opcode()) {
        is_invalid = false;
        match op {
            OpCode::EXTCODESIZE | OpCode::EXTCODEHASH | OpCode::EXTCODECOPY  => {
                let data = interp.stack.data();
                if data.len() >= 1 {
                    let addr_word = data[data.len() - 1];
                    let addr = Address::from_word(addr_word.into());
                    let (bytecode, _) = context.inner.code(addr).unwrap();
                    let data = bytecode.bytecode().clone();
                    unsafe {
                        GET_CODE_ADDRESS = addr;
                        GET_CODE_DATA = data;
                    }
                }
            },
            OpCode::CALL | OpCode::CALLCODE | OpCode::DELEGATECALL | OpCode::STATICCALL=> {
                let data = interp.stack.data();
                if data.len() >= 2 {
                    let addr_word = data[data.len() - 2];
                    let addr = Address::from_word(addr_word.into());
                    let (bytecode, _) = context.inner.code(addr).unwrap();
                    let data = bytecode.bytecode().clone();
                    unsafe {
                        GET_CODE_ADDRESS = addr;
                        GET_CODE_DATA = data;
                    }
                }
            },
            OpCode::CREATE => {
                let address = interp.contract.target_address;
                let account = context.inner.journaled_state.state.get(&address).unwrap();
                let nonce = account.info.nonce;

                unsafe {
                    GET_NONCE_ADDRESS = address;
                    GET_NONCE_NONCE = nonce;
                }
            },
            _ => (),
        }
    } else {
        is_invalid = true;
    }

    let mut mem_copy = vec![0; interp.shared_memory.len()];
    mem_copy.clone_from_slice(interp.shared_memory.context_memory());
    
    unsafe {
        HandleOpcode(
            stack_to_cstack(interp.stack.data()),
            bytes_to_csizedarray(&mem_copy.into()),
            address_to_caddress(interp.contract.target_address),
            interp.program_counter() as u64,
            interp.current_opcode(),
            is_invalid,
            context.inner.error.is_err(),
        );
    }
}

pub(crate) fn dep_call<DB:DatabaseExt>(data: &mut DepData, context: &mut EvmContext<DB>, inputs: &mut CallInputs) {
    on_enter(data, context, false, &inputs.input)
}

pub(crate) fn dep_call_end<DB:DatabaseExt>(data: &mut DepData, context: &mut EvmContext<DB>, _inputs: &CallInputs, outcome: &CallOutcome) {
    on_exit(data, context, &outcome.result)
}

pub(crate) fn dep_create<DB:DatabaseExt>(data: &mut DepData, context: &mut EvmContext<DB>, inputs: &mut CreateInputs) {
    on_enter(data, context, true, &inputs.init_code)
}

pub(crate) fn dep_create_end<DB:DatabaseExt>(data: &mut DepData, context: &mut EvmContext<DB>, _inputs: &CreateInputs, outcome: &CreateOutcome) {
    on_exit(data, context, &outcome.result)
}