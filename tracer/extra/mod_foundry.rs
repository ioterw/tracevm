use std::ptr;
use std::ffi::CString;

#[repr(C)]
struct Address {
    data: [u8; 20],
}
#[repr(C)]
struct Hash {
    data: [u8; 32],
}
#[repr(C)]
struct SizedArray {
    data: *const u8,
    size: i32,
}
#[repr(C)]
struct Stack {
    data: *const Hash,
    size: i32,
}

extern "C" {
    fn RegisterGetNonce(ptr: *const u8);
    fn RegisterGetCode(ptr: *const u8);

    fn InitDep(cfg: *const i8);

    fn StartTransactionRecording(
        isCreate: bool, addr: Address, input: SizedArray, block: u64,
        timestamp: u64, origin: Address, txHash: Hash,
        code: SizedArray, isSelfdestruct6780: bool, isRandom: bool,
    );
    fn EndTransactionRecording();

    fn HandleOpcode(
        stack: Stack, memory: SizedArray, addr: Address,
        pc: u64, op: u8, isInvalid: bool, hasError: bool,
    );
    fn HandleEnter(to: Address, input: SizedArray);
    fn HandleFault(op: u8);
    fn HandleExit(output: SizedArray, hasError: bool);
}

#[derive(Clone, Debug)]
pub(crate) struct DepData {
    call_depth: i32,
}
impl Default for DepData {
    fn default() -> DepData {
        let cfg = CString::new(
            "{\"kv\": {\"engine\": \"memory\", \"root\": \"\"}, \"logger\": {\"opcodes_short\": [\"e0\", \"e1\"], \"opcodes\": [], \"final_slots_short\": true, \"final_slots\": true, \"codes_short\": false, \"codes\": false, \"return_data_short\": false, \"return_data\": true, \"logs_short\": false, \"logs\": true, \"sol_view\": true}, \"output\": \"http://0.0.0.0:4334\", \"past_unknown\": false}"
        ).expect("CString::new failed");
        unsafe {
            InitDep(cfg.as_ptr());
        }
        DepData {
            call_depth: 0,
        }
    }
}

extern "C"
fn get_nonce(_addr: Address) -> u64 {
    0
}

extern "C"
fn get_code(_addr: Address) -> SizedArray {
    SizedArray{
        data: ptr::null(),
        size: 0,
    }
}

fn on_enter<EvmContext,CallInputs>(data: &mut DepData, _ecx: &mut EvmContext, _inputs: &mut CallInputs, _is_create: bool) {
    println!("HI enter");

    if data.call_depth == 0 {

    }

    data.call_depth += 1;
}

fn on_exit<EvmContext,Inputs>(data: &mut DepData, _context: &mut EvmContext, _inputs: &Inputs, _is_create: bool) {
    println!("HI exit");

    data.call_depth -= 1;
}

pub(crate) fn dep_step<Interpreter,EvmContext>(_data: &mut DepData, _interp: &mut Interpreter, _ecx: &mut EvmContext) {
    println!("HI step")
}

pub(crate) fn dep_call<EvmContext,CallInputs>(data: &mut DepData, ecx: &mut EvmContext, inputs: &mut CallInputs) {
    on_enter(data, ecx, inputs, false)
}

pub(crate) fn dep_call_end<EvmContext,CallInputs>(data: &mut DepData, context: &mut EvmContext, inputs: &CallInputs) {
    on_exit(data, context, inputs, false)
}

pub(crate) fn dep_create<EvmContext,CreateInputs>(data: &mut DepData, ecx: &mut EvmContext, inputs: &mut CreateInputs) {
    on_enter(data, ecx, inputs, true)
}

pub(crate) fn dep_create_end<EvmContext,CreateInputs>(data: &mut DepData, context: &mut EvmContext, inputs: &CreateInputs) {
    on_exit(data, context, inputs, true)
}
