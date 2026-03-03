import sys, pyvisa, warnings
warnings.filterwarnings('ignore')

if len(sys.argv) < 2:
    print("ERROR: no command"); sys.exit(1)
cmd = sys.argv[1]

GEN = 'USB0::0x6656::0x0834::AWG1224500035::INSTR'

try:
    rm    = pyvisa.ResourceManager()        
    instr = rm.open_resource(GEN)
    instr.timeout           = 8000
    instr.write_termination = '\n'
    instr.read_termination  = ''         

    if cmd.strip() == '*IDN?':
        print(instr.query('*IDN?').strip())
    elif '?' in cmd:
        print("QUERY_NOT_SUPPORTED")
    else:
        instr.write(cmd)
        print("OK")

    instr.close()
    rm.close()
except Exception as e:
    print(f"ERROR: {e}"); sys.exit(1)
