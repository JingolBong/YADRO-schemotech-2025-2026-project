import sys
import pyvisa

cmd = " ".join(sys.argv[1:])
try:
    rm = pyvisa.ResourceManager()
    resources = rm.list_resources()
    gen_id = next((r for r in resources if "AWG" in r), resources[0] if resources else "")
    
    gen = rm.open_resource(gen_id)
    gen.write(cmd)
    gen.close()
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)