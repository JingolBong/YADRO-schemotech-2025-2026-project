import sys
import os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../..')))
sys.path.insert(0, os.path.dirname(__file__))
import grpc
from concurrent import futures
import time
import pyvisa
import re
import math
import lab_pb2
import lab_pb2_grpc

def parse_val(text):
    if not text or "ERROR" in text or "NAN" in text: return 0.0
    m = re.search(r'([-+]?[0-9]*\.?[0-9]+(?:[eE][-+]?[0-9]+)?)', text)
    if not m: return 0.0
    val = float(m.group(1))
    if 'mV' in text: val /= 1000.0
    elif 'uV' in text: val /= 1e6
    return val

class InstrumentControllerServicer(lab_pb2_grpc.InstrumentControllerServicer):
    def __init__(self):
        self.mock_mode = False
        try:
            self.rm = pyvisa.ResourceManager()
            resources = self.rm.list_resources()
            if len(resources) < 2:
                raise ValueError("Недостаточно приборов")
                
            self.gen = self.rm.open_resource(next((r for r in resources if "AWG" in r), resources[0]))
            self.osc = self.rm.open_resource(next((r for r in resources if "AWG" not in r), resources[-1]))
            self.osc.timeout = 10000
            print("✅ Python gRPC Server: Приборы подключены! (ПРОДАКШЕН РЕЖИМ)")
        except Exception as e:
            self.mock_mode = True
            print(f"⚠️ Железо не найдено ({e}). Сервер запущен в РЕЖИМЕ СИМУЛЯЦИИ (Mock Mode)!")

    def SetupInstruments(self, request, context):
        if self.mock_mode:
            return lab_pb2.SetupResponse(success=True)
            
        try:
            self.gen.write(f":CHANnel1:BASE:WAVe {request.waveform}")
            self.gen.write(f":CHANnel1:BASE:AMPLitude {request.amplitude_vpp}")
            self.gen.write(":CHANnel1:OUTPut ON")
            
            self.osc.write(":CH1:DISPlay ON")
            self.osc.write(":CH2:DISPlay ON")
            self.osc.write(":MEASUrement:MEAS1:SOUrce CH1")
            self.osc.write(":MEASUrement:MEAS1:TYPe PKPK")
            self.osc.write(":MEASUrement:MEAS2:SOUrce CH2")
            self.osc.write(":MEASUrement:MEAS2:TYPe PKPK")
            self.osc.write(":MEASUrement:MEAS3:TYPe RPHAse")
            return lab_pb2.SetupResponse(success=True)
        except Exception as e:
            return lab_pb2.SetupResponse(success=False, error_msg=str(e))

    def SetGenerator(self, request, context):
        if self.mock_mode:
            return lab_pb2.GeneratorResponse(success=True)
            
        if request.frequency_hz > 0:
            self.gen.write(f":CHANnel1:BASE:FREQuency {int(request.frequency_hz)}")
        if request.amplitude_vpp > 0:
            self.gen.write(f":CHANnel1:BASE:AMPLitude {request.amplitude_vpp}")
        return lab_pb2.GeneratorResponse(success=True)

    def MeasurePoint(self, request, context):
        if self.mock_mode:
            time.sleep(0.05) 
            f = request.frequency_hz
            fc = 1000.0 
            vin = 4.0
            vout = vin / math.sqrt(1 + (f/fc)**2)
            phase = -math.atan(f/fc) * (180.0 / math.pi)
            import random
            vout += random.uniform(-0.02, 0.02)
            return lab_pb2.MeasureResponse(vin_vpp=vin, vout_vpp=vout, phase_shift_deg=phase)

        try:
            self.gen.write(f":CHANnel1:BASE:FREQuency {int(request.frequency_hz)}")
            self.osc.write(f":HORIzontal:SCALe {request.timebase}")
            self.osc.write(f":CH2:SCALe {request.vscale}")
            
            time.sleep(2.5)
            
            vin_raw = self.osc.query(":MEASUrement:MEAS1:VALue?").strip()
            vout_raw = self.osc.query(":MEASUrement:MEAS2:VALue?").strip()
            phase_raw = self.osc.query(":MEASUrement:MEAS3:VALue?").strip()
            
            vin = parse_val(vin_raw)
            if vin < 0.1: vin = 4.0 
            
            return lab_pb2.MeasureResponse(
                vin_vpp=vin, 
                vout_vpp=parse_val(vout_raw),
                phase_shift_deg=parse_val(phase_raw)
            )
        except pyvisa.errors.VisaIOError as e:
            return lab_pb2.MeasureResponse(error_msg=f"Timeout: {e}")

    def Shutdown(self, request, context):
        if self.mock_mode:
            return lab_pb2.ShutdownResponse(success=True)
        self.gen.write(":CHANnel1:OUTPut OFF")
        return lab_pb2.ShutdownResponse(success=True)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    lab_pb2_grpc.add_InstrumentControllerServicer_to_server(InstrumentControllerServicer(), server)
    server.add_insecure_port('[::]:50051')
    print("🚀 Server is running on port 50051...")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()