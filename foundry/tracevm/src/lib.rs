use std::{ ffi::{CString, c_char} };
use queues::{ Queue, IsQueue, queue };
use revm::{
    EvmContext, Database,
    interpreter::{ Interpreter, OpCode, CallInputs, CallOutcome, CreateInputs, CreateOutcome, InterpreterResult },
    primitives::{ SpecId },
};
use alloy_primitives::{ address, Address, Bytes, U256, FixedBytes, Log };
use alloy_sol_types::{ sol, SolEvent };
use serde::{Deserialize};
use hex::FromHex;

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

    fn InitDep(cfg: *const i8, ptr: Option<extern "C" fn(*const c_char)>);

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
    fn HandleFault(op: u8);
}

#[repr(u8)]
pub enum DepDataType {
    Debug = 1,
    Trace = 2,
}

#[derive(Clone, Debug)]
pub struct DepData<const DATA_TYPE: u8> {
    pub call_depth: i32,
    pub activated: bool,
}
impl<const DATA_TYPE: u8> DepData<DATA_TYPE> {
    pub fn clear(&mut self) {
        panic!("I don't know what to do with this");
    }
}
impl<const DATA_TYPE: u8> Default for DepData<DATA_TYPE> {
    fn default() -> DepData<DATA_TYPE> {
        unsafe {
            // easy fix
            if DEP_DATA_TYPE != 0 {
                if DEP_DATA_TYPE == DepDataType::Debug as u8 && DATA_TYPE == DepDataType::Trace as u8 {
                    return DepData {
                        call_depth: 0,
                        activated: false,
                    }
                } else {
                    panic!("Unknown DEP_DATA_TYPEs: {}, {}", DEP_DATA_TYPE, DATA_TYPE);
                }
            }
            DEP_DATA_TYPE = DATA_TYPE;
        }

        let cfg: &str;
        let callback: Option<extern "C" fn(*const c_char)>;

        match DATA_TYPE {
            1_u8 => {
                cfg = "{
                    \"kv\": {\"engine\": \"amnesia\", \"root\": \"\"}, 
                    \"logger\": {
                        \"opcodes_short\": [\"e0\", \"e1\", \"e2\", \"e3\"],
                        \"opcodes\": [],
                        \"final_slots_short\": true,
                        \"final_slots\": false,
                        \"codes_short\": false,
                        \"codes\": false,
                        \"return_data_short\": false,
                        \"return_data\": false,
                        \"logs_short\": false,
                        \"logs\": false,
                        \"sol_view\": true,
                        \"minimal_info\": true,
                        \"omit_info\": false,
                        \"omit_formulas\": false,
                        \"output_format\": \"text\"
                    },
                    \"output\": \"http://0.0.0.0:4334\",
                    \"past_unknown\": true
                }";
                callback = None;
            }
            2_u8 => {
                cfg = "{
                    \"kv\": {\"engine\": \"amnesia\", \"root\": \"\"}, 
                    \"logger\": {
                        \"opcodes_short\": [\"e0\", \"e1\", \"e2\", \"e3\"],
                        \"opcodes\": [],
                        \"final_slots_short\": false,
                        \"final_slots\": false,
                        \"codes_short\": false,
                        \"codes\": false,
                        \"return_data_short\": false,
                        \"return_data\": false,
                        \"logs_short\": false,
                        \"logs\": false,
                        \"sol_view\": true,
                        \"minimal_info\": false,
                        \"omit_info\": true,
                        \"omit_formulas\": true,
                        \"output_format\": \"json\"
                    },
                    \"output\": \"\",
                    \"past_unknown\": true
                }";
                callback = Some(trace_callback);
            }
            _ => panic!("Unknown DATA_TYPE")
        }

        let ccfg = CString::new(cfg).expect("CString::new failed");
        unsafe {
            InitDep(ccfg.as_ptr(), callback);
            RegisterGetNonce(get_nonce);
            RegisterGetCode(get_code);
        }
        DepData {
            call_depth: 0,
            activated: true,
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

extern "C"
fn trace_callback(data: *const c_char) {
    let c_str: &std::ffi::CStr;
    unsafe {
        c_str = std::ffi::CStr::from_ptr(data);
    }
    let str_slice: &str = c_str.to_str().unwrap();
    unsafe {
        TRACE_CALLBACK_DATA.push(str_slice);
    }
}
sol! {
    #[derive(Default, PartialEq, Debug)]
    event SLOAD(
        bytes32 key,
        bytes32 value,
        string[] offsets,
    );
    event SSTORE(
        bytes32 key,
        bytes32 value,
        string[] offsets,
    );
    event TLOAD(
        bytes32 key,
        bytes32 value,
        string[] offsets,
    );
    event TSTORE(
        bytes32 key,
        bytes32 value,
        string[] offsets,
    );
}
pub struct TraceCallbackData {
    queue: Option<Queue<String>>,
}
impl TraceCallbackData {
    fn push(&mut self, data: &str) {
        if self.queue.is_none() {
            self.queue = queue![].into();
        }
        let queue: &mut Queue<String> = self.queue.as_mut().unwrap();
        let _ = queue.add(data.to_string());
    }
    pub fn pull(&mut self) -> Option<Log> {
        if self.queue.is_none() {
            return None
        }
        let queue = self.queue.as_mut().unwrap();
        if queue.size() == 0 {
            return None
        }

        #[derive(Deserialize)]
        struct JSONSolidity {
            opcode: String,
            key: String,
            value: String,
            offsets: Vec<(String,String)>,
        }
        #[derive(Deserialize)]
        struct JSONInfo {
            solidity: JSONSolidity,
        }

        let json_data = queue.remove().unwrap();
        let json_info: JSONInfo = serde_json::from_str(json_data.as_str()).unwrap();

        let key_fixed = FixedBytes::<32>::from_hex(json_info.solidity.key).unwrap();
        let value_fixed = FixedBytes::<32>::from_hex(json_info.solidity.value).unwrap();
        let mut offsets_fixed: Vec<String> = Vec::with_capacity(json_info.solidity.offsets.len());
        for i in 0..json_info.solidity.offsets.len() {
            let (a, b) = &json_info.solidity.offsets[i];
            offsets_fixed.push(a.to_owned() + " " + &b);
        }

        let (event_data, event_topics) = match json_info.solidity.opcode.as_str() {
            "sstore" => {
                let event = SSTORE { key: key_fixed, value: value_fixed, offsets: offsets_fixed };
                (event.encode_data(), event.encode_topics())
            },
            "sload" => {
                let event = SLOAD { key: key_fixed, value: value_fixed, offsets: offsets_fixed };
                (event.encode_data(), event.encode_topics())
            },
            "tstore" => {
                let event = TSTORE { key: key_fixed, value: value_fixed, offsets: offsets_fixed };
                (event.encode_data(), event.encode_topics())
            },
            "tload" => {
                let event = TLOAD { key: key_fixed, value: value_fixed, offsets: offsets_fixed };
                (event.encode_data(), event.encode_topics())
            },
            _ => panic!("Unknown opcode"),
        };

        let event_topics_fixed: Vec<FixedBytes<32>> = event_topics.iter().map(|&i|i.into()).collect();
        let log = Log::new(
            address!("0000000000000000000000000000000000000000"),
            event_topics_fixed,
            event_data.into(),
        ).expect("Log generation failed");
        Some(log)
    }
}
pub static mut TRACE_CALLBACK_DATA: TraceCallbackData = TraceCallbackData{
    queue: None,
};

static mut ACTIVATED_HASH: FixedBytes<32> = FixedBytes::ZERO;
static mut DEP_DATA_TYPE: u8 = 0;
fn is_activated<const DATA_TYPE: u8>(data: &DepData<DATA_TYPE>) -> bool {
    if !data.activated {
        return false;
    }
    unsafe {
        return ACTIVATED_HASH != FixedBytes::ZERO;
    }
}
pub fn activate(hash: FixedBytes<32>) {
    unsafe {
        ACTIVATED_HASH = hash;
    }
}

fn on_enter<DB:Database, const DATA_TYPE: u8>(data: &mut DepData<DATA_TYPE>, context: &mut EvmContext<DB>, is_create: bool, input: &Bytes, addr: Address) {
    if !is_activated(data) {
        return;
    }

    if data.call_depth == 0 {
        let input = context.inner.env.tx.data.clone();
        let origin = context.inner.env.tx.caller;
        let code: Bytes;
        if is_create {
            code = Bytes::new()
        } else {
            let bytecode = match context.code(addr) {
                Ok((bytecode, _)) => bytecode,
                Err(_) => panic!("context.code(addr) failed"),
            };
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
                CHash{ data: *ACTIVATED_HASH },
                bytes_to_csizedarray(&code),
                is_selfdestruct6780,
                is_random,
            );
        }
    }

    unsafe {
        HandleEnter(
            address_to_caddress(addr),
            bytes_to_csizedarray(&input),
        )
    }

    data.call_depth += 1;
}

fn on_exit<DB:Database, const DATA_TYPE: u8>(data: &mut DepData<DATA_TYPE>, context: &mut EvmContext<DB>, result: &InterpreterResult) {
    if !is_activated(data) {
        return;
    }

    data.call_depth -= 1;

    unsafe {
        HandleExit(
            bytes_to_csizedarray(&result.output),
            context.inner.error.is_err(),
        )
    }

    if data.call_depth == 0 {
        unsafe {
            EndTransactionRecording();
        }
    }
}

pub fn dep_step<DB:Database, const DATA_TYPE: u8>(data: &mut DepData<DATA_TYPE>, interp: &mut Interpreter, context: &mut EvmContext<DB>) {
    if !is_activated(data) {
        return;
    }

    let is_invalid: bool;
    if let Some(op) = OpCode::new(interp.current_opcode()) {
        is_invalid = false;
        match op {
            OpCode::EXTCODESIZE | OpCode::EXTCODEHASH | OpCode::EXTCODECOPY  => {
                let data = interp.stack.data();
                if data.len() >= 1 {
                    let addr_word = data[data.len() - 1];
                    let addr = Address::from_word(addr_word.into());
                    let bytecode = match context.inner.code(addr) {
                        Ok((bytecode, _)) => bytecode,
                        Err(_) => panic!("context.inner.code(addr) failed"),
                    };
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
                    let bytecode = match context.inner.code(addr) {
                        Ok((bytecode, _)) => bytecode,
                        Err(_) => panic!("context.inner.code(addr) failed"),
                    };
                    let data = bytecode.bytecode().clone();
                    unsafe {
                        GET_CODE_ADDRESS = addr;
                        GET_CODE_DATA = data;
                    }
                }
            },
            OpCode::CREATE => {
                let address = interp.contract.target_address;
                let nonce = context.journaled_state.account(address).info.nonce;

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

pub fn dep_step_end<DB:Database, const DATA_TYPE: u8>(data: &mut DepData<DATA_TYPE>, interp: &mut Interpreter, context: &mut EvmContext<DB>) {
    if !is_activated(data) {
        return;
    }

    if context.inner.error.is_err() {
        unsafe {
            HandleFault(interp.current_opcode())
        }
    }
}

pub fn dep_call<DB:Database, const DATA_TYPE: u8>(data: &mut DepData<DATA_TYPE>, context: &mut EvmContext<DB>, inputs: &mut CallInputs) {
    let addr = inputs.target_address;
    on_enter(data, context, false, &inputs.input, addr)
}

pub fn dep_call_end<DB:Database, const DATA_TYPE: u8>(data: &mut DepData<DATA_TYPE>, context: &mut EvmContext<DB>, _inputs: &CallInputs, outcome: &CallOutcome) {
    on_exit(data, context, &outcome.result)
}

pub fn dep_create<DB:Database, const DATA_TYPE: u8>(data: &mut DepData<DATA_TYPE>, context: &mut EvmContext<DB>, inputs: &mut CreateInputs) {
    let addr = inputs.created_address(context.journaled_state.account(inputs.caller).info.nonce);
    on_enter(data, context, true, &inputs.init_code, addr)
}

pub fn dep_create_end<DB:Database, const DATA_TYPE: u8>(data: &mut DepData<DATA_TYPE>, context: &mut EvmContext<DB>, _inputs: &CreateInputs, outcome: &CreateOutcome) {
    on_exit(data, context, &outcome.result)
}
