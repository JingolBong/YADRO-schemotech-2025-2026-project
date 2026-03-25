import sys, time, pyvisa, warnings
warnings.filterwarnings('ignore')

if len(sys.argv) < 2:
    print("ERROR: no command"); sys.exit(1)
cmd = " ".join(sys.argv[1:])

OSC = 'USB0::0x5345::0x1235::24410241::INSTR'

try:
    rm = pyvisa.ResourceManager()
    instr = rm.open_resource(OSC)
    
    instr.timeout = 5000
    instr.write_termination = '\n'
    instr.read_termination = ''

    if '?' in cmd:
        instr.clear()
        time.sleep(0.1)
        instr.write(cmd)
        time.sleep(0.3)
        try:
            raw = instr.read_raw(size=4096)
            resp = raw.decode('utf-8', errors='ignore').strip().rstrip('->').strip()
            print(resp if resp else "EMPTY")
        except Exception as e:
            print(f"ERROR_READ: {e}")
    else:
        instr.write(cmd)
        time.sleep(0.05)
        print("OK")

    instr.close()
    rm.close()
except Exception as e:
    print(f"ERROR: {e}"); sys.exit(1)