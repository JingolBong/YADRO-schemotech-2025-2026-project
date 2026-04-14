import sys
import pyvisa

cmd = " ".join(sys.argv[1:])
try:
    rm = pyvisa.ResourceManager()
    resources = rm.list_resources()
    osc_id = next((r for r in resources if "AWG" not in r), resources[-1] if resources else "")

    osc = rm.open_resource(osc_id)
    osc.timeout = 10000  # Драйвер теперь правильный, но таймаут оставляем с запасом
    
    if "?" in cmd:
        print(osc.query(cmd).strip()) # Печатаем ТОЛЬКО цифру
    else:
        osc.write(cmd)
        
    osc.close()
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)