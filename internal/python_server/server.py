import sys
import os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "../..")))
sys.path.insert(0, os.path.dirname(__file__))

import grpc
from concurrent import futures
import time
import pyvisa
import re
import math
import random
import numpy as np
import lab_pb2
import lab_pb2_grpc

def parse_val(text):
    if not text: return 0.0
    text = text.strip()
    if "ERROR" in text or "NAN" in text or "?" in text: return 0.0
    
    m = re.search(r'([-+]?[0-9]*\.?[0-9]+(?:[eE][-+]?[0-9]+)?)', text)
    if not m: return 0.0
    val = float(m.group(1))

    if "mV" in text or "mv" in text: val /= 1000.0
    elif "uV" in text or "uv" in text: val /= 1e6
    
    return val

def parse_timebase(tb_str):
    tb_str = tb_str.lower().strip()
    if 'ms' in tb_str: return float(tb_str.replace('ms', '')) * 1e-3
    if 'us' in tb_str: return float(tb_str.replace('us', '')) * 1e-6
    if 'ns' in tb_str: return float(tb_str.replace('ns', '')) * 1e-9
    if 's' in tb_str: return float(tb_str.replace('s', ''))
    return 1.0

def calc_phase_by_fft(raw1, raw2):
    """Вычисляет фазу через Быстрое преобразование Фурье (FFT)"""
    if len(raw1) < 2000 or len(raw2) < 2000:
        return 0.0

    start = 200
    end = min(len(raw1), len(raw2)) - 200
    if (end - start) % 2 != 0: 
        end -= 1
        
    chunk1 = raw1[start:end]
    chunk2 = raw2[start:end]

    y1 = np.frombuffer(chunk1, dtype='<i2').astype(np.float32)
    y2 = np.frombuffer(chunk2, dtype='<i2').astype(np.float32)

    y1 -= np.mean(y1)
    y2 -= np.mean(y2)

    Y1 = np.fft.rfft(y1)
    Y2 = np.fft.rfft(y2)

    idx = np.argmax(np.abs(Y1[1:])) + 1

    phase1 = np.angle(Y1[idx])
    phase2 = np.angle(Y2[idx])

    phase_diff = np.degrees(phase2 - phase1)
    phase_diff = (phase_diff + 180) % 360 - 180
    
    return phase_diff

class InstrumentControllerServicer(lab_pb2_grpc.InstrumentControllerServicer):
    def __init__(self):
        self.mock_mode = False
        try:
            self.rm = pyvisa.ResourceManager()
            resources = self.rm.list_resources()
            print(f"resources={resources}", flush=True)
            if len(resources) < 2:
                raise ValueError("not enough instruments")
            
            self.gen = self.rm.open_resource(next((r for r in resources if "AWG" in r), resources[0]))
            self.osc = self.rm.open_resource(next((r for r in resources if "AWG" not in r), resources[-1]))
            
            self.osc.timeout = 3000
            self.osc.read_termination = '\n' 
            self.osc.write_termination = '\n'
            
            try: self.osc.write("*CLS") 
            except: pass

            print("server started in real mode", flush=True)
        except Exception as e:
            self.mock_mode = True
            print(f"server started in mock mode: {e}", flush=True)

    def SetupInstruments(self, request, context):
        if self.mock_mode: return lab_pb2.SetupResponse(success=True)
        try:
            print(f"SetupInstruments amp={request.amplitude_vpp} wave={request.waveform}", flush=True)
            
            self.gen.write(f":CHANnel1:BASE:WAVe {request.waveform}")
            self.gen.write(f":CHANnel1:BASE:AMPLitude {request.amplitude_vpp}")
            self.gen.write(":CHANnel1:BASE:OFFSet 0")
            self.gen.write(":CHANnel1:OUTPut ON")

            self.osc.write(":CH1:DISPlay ON")
            self.osc.write(":CH2:DISPlay ON")
            self.osc.write(":CH1:PROBe 1X")
            self.osc.write(":CH2:PROBe 1X")
            
            time.sleep(0.5)
            return lab_pb2.SetupResponse(success=True)
        except Exception as e:
            print(f"Setup error={e}", flush=True)
            return lab_pb2.SetupResponse(success=False, error_msg=str(e))

    def SetGenerator(self, request, context):
        if self.mock_mode: return lab_pb2.GeneratorResponse(success=True)
        try:
            if request.frequency_hz > 0:
                self.gen.write(f":CHANnel1:BASE:FREQuency {int(request.frequency_hz)}")
            if request.amplitude_vpp > 0:
                self.gen.write(f":CHANnel1:BASE:AMPLitude {request.amplitude_vpp}")
            return lab_pb2.GeneratorResponse(success=True)
        except Exception as e:
            return lab_pb2.GeneratorResponse(success=False)

    def get_measurement(self, source, meas_type):
        """ Текстовый запрос параметров (включаем терминатор) """
        self.osc.read_termination = '\n'
        queries = [
            f":MEASUrement:{source}:{meas_type}?", 
            f":MEASure:{meas_type}? {source}",
            f":MEASUrement:MEAS1:SOUrce {source}; :MEASUrement:MEAS1:TYPe {meas_type}; :MEASUrement:MEAS1:VALue?",
            f":MEASure[1]:{meas_type}?" if source=="CH1" else f":MEASure[2]:{meas_type}?"
        ]
        
        for q in queries:
            try:
                self.osc.write("*CLS")
                res = self.osc.query(q).strip()
                if res and "ERROR" not in res and res != "":
                    return res
            except Exception:
                pass
            time.sleep(0.05) 
        return ""

    def get_raw_waveform(self, channel):
        """ Запрос бинарного массива (отключаем терминатор) """
        self.osc.read_termination = None 
        try: self.osc.clear()
        except: pass
        
        try:
            try: 
                self.osc.write(":DATA:WAVE:SCREen:HEAD?")
                time.sleep(0.1)
                self.osc.read_raw() 
            except: pass

            self.osc.write(f":DATA:WAVE:SCREen:{channel}?")
            time.sleep(0.2)
            raw = self.osc.read_raw()
            
            self.osc.read_termination = '\n'
            return raw
        except Exception as e:
            self.osc.read_termination = '\n'
            print(f"ОШИБКА загрузки {channel}: {e}", flush=True)
            return b""

    def MeasurePoint(self, request, context):
        if self.mock_mode:
            time.sleep(0.05)
            f = request.frequency_hz
            fc = 1000.0
            vin = 4.0
            vout = vin / math.sqrt(1 + (f / fc) ** 2)
            phase = -math.atan(f / fc) * (180.0 / math.pi)
            vout += random.uniform(-0.02, 0.02)
            return lab_pb2.MeasureResponse(vin_vpp=vin, vout_vpp=vout, phase_shift_deg=phase)

        try:
            self.gen.write(f":CHANnel1:BASE:FREQuency {int(request.frequency_hz)}")
            self.osc.write(f":HORIzontal:SCALe {request.timebase}")
            self.osc.write(f":CH2:SCALe {request.vscale}")
            
            time.sleep(0.3) 
            
            vin_raw = self.get_measurement("CH1", "PKPK")
            vout_raw = self.get_measurement("CH2", "PKPK")
            
            raw_ch1 = self.get_raw_waveform("CH1")
            raw_ch2 = self.get_raw_waveform("CH2")

            phase_deg = calc_phase_by_fft(raw_ch1, raw_ch2)

            vin = parse_val(vin_raw)
            vout = parse_val(vout_raw)

            print(f"Measure f={request.frequency_hz:.1f} | Vin={vin:.3f} | Vout={vout:.3f} | Phase={phase_deg:.1f}°", flush=True)

            if vin < 0.1:
                vin = request.amplitude_vpp if request.amplitude_vpp > 0 else 4.0

            return lab_pb2.MeasureResponse(
                vin_vpp=vin,
                vout_vpp=vout,
                phase_shift_deg=phase_deg,
            )
        except Exception as e:
            print(f"Measure error={e}", flush=True)
            return lab_pb2.MeasureResponse(error_msg=str(e))

    def Shutdown(self, request, context):
        if not self.mock_mode:
            try: self.gen.write(":CHANnel1:OUTPut OFF")
            except: pass
        return lab_pb2.ShutdownResponse(success=True)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    lab_pb2_grpc.add_InstrumentControllerServicer_to_server(InstrumentControllerServicer(), server)
    server.add_insecure_port("[::]:50051")
    print("grpc listening on :50051", flush=True)
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()