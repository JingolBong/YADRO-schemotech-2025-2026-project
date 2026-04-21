import sys
import pyvisa

cmd = " ".join(sys.argv[1:])
try:
    rm = pyvisa.ResourceManager()
    resources = rm.list_resources()
    osc_id = next((r for r in resources if "AWG" not in r), resources[-1] if resources else "")

    osc = rm.open_resource(osc_id)
    osc.timeout = 10000 
    
    if "?" in cmd:
        print(osc.query(cmd).strip()) 
    else:
        osc.write(cmd)
        
    osc.close()
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)