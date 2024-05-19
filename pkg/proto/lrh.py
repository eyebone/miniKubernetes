# -*- coding: utf-8 -*-
"""
Create Time: 2020/9/17 10:48
Author: Liuqing Zhang
"""
import datetime
import os
import shutil
import time

from PyQt5 import QtCore
from PyQt5 import QtWidgets
from PyQt5.QtCore import pyqtSignal, QTimer, Qt
from PyQt5.QtWidgets import QMainWindow, QApplication, QVBoxLayout, QMessageBox, QTableWidgetItem
from sys import argv, exit
import pandas as pd
import numpy as np
from pickle import dump, load
import pyqtgraph.console
import pyqtgraph as pg
from ui.demo_for_stamping_six_pip import Ui_MainWindow
import threading
# from sdk.daq_sdk import DaqSDK
from sdk.daq_sdk import *
import run_core
import scipy
from scipy import signal

from model_predict import PredictMain as predict
import multiprocessing as mp


class ton_class(QtCore.QObject):
    signal_error = pyqtSignal()


    def __init__(self, ton_curve=None, ton_curve_win=None, ton_channels=[], SG_list=[], fs=None):
        super(ton_class, self).__init__()
        self.index = 0
        self.ton_curve = ton_curve
        self.ton_curve_win = ton_curve_win
        self.num_channel = len(ton_channels)
        # self.ton_channels = [int(channel[7]) for channel in ton_channels]
        self.ton_channels = ton_channels
        self.ton_range = [[0, 9999] for channel in range(6 + 1)]
        # print(self.ton_channels)
        # assert (self.num_channel == 4)
        self.buffer_result = [[] for i in range(6)]  # 缓存传感器求出的吨位结果
        self.total_ton = []  # 缓存总吨位
        self.is_plotted = 0
        self.fs = fs
        self.curve_color = [(255, 0, 0), (0, 255, 0), (0, 0, 255), (128, 128, 128), (0, 0, 128), (0, 128, 0)]
        # self.ton_plot([[i] for i in range(6)], 1)
        self.output_labels = [None for channel in range(6 + 1)]  # 各通道各一个，总和一个
        self.spinBox_ton_calibration = [None for channel in range(6 + 1)]  # 各通道一个spinBox，再加一个按钮
        self.ton_calibration = [1 for channel in range(6)]
        self.spinBox_ton_real = [None for channel in range(6 + 1)]  # 各通道一个spinBox，再加一个按钮
        self.calibration_time = 1
        self.auto_load_calibration()
        self.auto_save_calibration()
        self.SG_list = SG_list
        # self.ton_curve.enableAutoRange()

    def set_output_label(self, output_labels):
        for channel in self.ton_channels:
            self.output_labels[channel] = output_labels[channel]
        self.output_labels[-1] = output_labels[-1]

    def set_calibration_time(self, calibration_time):
        self.calibration_time = calibration_time

    def set_ton_range(self, ton_range):
        for channel in self.ton_channels:
            self.ton_range[channel] = [ton_range[channel][0].value(), ton_range[channel][1].value()]
        self.ton_range[-1] = [ton_range[-1][0].value(), ton_range[-1][1].value()]

    def set_spinBox_ton_real(self, spinBox_ton_real):
        self.spinBox_ton_real = spinBox_ton_real

    def set_spinBox_ton_calibration(self, ton_calibration):
        self.spinBox_ton_calibration = ton_calibration
        self.update_ton_calibration_display()

    def update_ton_calibration_value(self):
        for channel in self.ton_channels:
            self.ton_calibration[channel] = self.spinBox_ton_calibration[channel].value()
        self.auto_save_calibration()

    def update_ton_calibration_display(self):
        for channel in self.ton_channels:
            self.spinBox_ton_calibration[channel].setValue(self.ton_calibration[channel])

    def calculate_calibration(self):
        for channel in self.ton_channels:
            calculated_value = np.average(self.buffer_result[channel][-self.calibration_time:])
            real_value = self.spinBox_ton_real[channel].value()
            self.ton_calibration[channel] *= real_value / calculated_value
        self.auto_save_calibration()
        self.update_ton_calibration_display()

    def change_calibration(self, channel):
        # print('channel' + str(channel) + ' ready')
        self.spinBox_ton_calibration[channel].setEnabled(True)
        self.spinBox_ton_real[channel].setEnabled(True)
        if all([self.spinBox_ton_calibration[channel_index].isEnabled() for channel_index in self.ton_channels]):
            self.spinBox_ton_calibration[-1].setEnabled(True)
        if all([self.spinBox_ton_real[channel_index].isEnabled() for channel_index in self.ton_channels]):
            self.spinBox_ton_real[-1].setEnabled(True)

    def reset_calibration(self):
        self.ton_calibration = [1 for channel in range(6)]
        self.update_ton_calibration_display()
        self.auto_save_calibration()

    def auto_save_calibration(self, filename='标定系数.txt'):
        # 将标定系数以txt格式保存
        # print(self.ton_calibration)
        np.savetxt(filename, self.ton_calibration)

    def auto_load_calibration(self, filename='标定系数.txt'):
        try:
            self.ton_calibration = np.loadtxt(filename)
            # print(self.ton_calibration)
        except IOError:
            print('没有找到标定系数文件')

    def get_ton_value(self, length):
        if len(self.buffer_result[0]) > 0:
            return self.buffer_result[:, 0]

    def receive_data(self, channel, data, stamping_counter):  # 接受单个通道数据的槽函数
        # print('receive data')
        # print(len(data))
        if channel in self.ton_channels:
            # channel_index = self.ton_channels.index(channel)
            if self.SG_list[channel]:
                ton_value = self.calculate_SG(data=data) * self.ton_calibration[channel]
            else:
                ton_value = self.calculate(data=data) * self.ton_calibration[channel]
            self.buffer_result[channel].append(ton_value)
            self.output_labels[channel].setText('%.3f' % ton_value)
            if self.is_plotted < stamping_counter and all(
                    [len(self.buffer_result[channel]) >= stamping_counter for channel in self.ton_channels]):
                self.is_plotted = stamping_counter
                self.output_labels[-1].setText(
                    '%.3f' % np.sum(
                        [self.buffer_result[channel][stamping_counter - 1] for channel in self.ton_channels]))
                self.ton_plot(self.buffer_result, stamping_counter)

    def calculate(self, data=None, fs=1000, filter_fre=50):
        # self.ton_curve.plot().setData([self.index], np.random.random(1))
        # QApplication.processEvents()
        dT = 1 / fs
        len_data = len(data)
        # cut_length = len_data // 2
        N = int(2 ** np.ceil(np.log2(len_data) + 1))
        f = [i * fs / N for i in range(N)]
        data_f = np.fft.fft(data, N) / N
        # np.savetxt('a.txt', data_f)
        cut_length = 500
        win_s = int(np.round(filter_fre * N / fs))
        TuWIN = signal.tukey(2 * cut_length, 1.0)
        TuWIN1 = TuWIN[cut_length - 1:]
        Win = np.concatenate((np.ones(win_s), TuWIN1, np.zeros(N - win_s - cut_length - 1)))
        # data_f = np.pad(data_f, (0, len(Win) - N), 'constant', constant_values=(0, 0))
        data_f = data_f * Win
        data_r = np.fft.ifft(data_f) * N * 2
        res_real = np.real(data_r)
        res_real = res_real[0: len_data]
        temp = res_real
        #         for i in range(len_data - 1):
        #             temp[i] = res_real[i] + res_real[i + 1]
        result = np.max(temp)
        if result > 0:
            return result
        else:
            return 0
        # return np.max(temp) / 2

    def calculate_SG(self, data=None):
        temp = data
        result = np.max(np.abs(temp))
        if result > 0:
            return result
        else:
            return 0

    def save_ton_data(self, stamping_counter=-1):
        filename = "tondata//" + datetime.datetime.now().strftime("%Y%m%d_%H%M_%S") + ".csv"
        column_name = ['通道',  '标定系数', '实际吨位', '报警吨位']
        column_channel = ['channel' + str(channel) for channel in self.ton_channels] + ['total']
        column_calibration = [self.ton_calibration[channel] for channel in self.ton_channels] + ['\\']
        original_data = [self.buffer_result[channel][stamping_counter] for channel in self.ton_channels]
        column_real = [data for data in original_data] + ['\\']
        column_range = [str(self.ton_range[channel][0]) + ' - ' + str(self.ton_range[channel][1]) for channel in (self.ton_channels + [-1])]
        dataframe = pd.DataFrame({column_name[0]: column_channel, column_name[1]: column_calibration, column_name[2]: column_real, column_name[3]: column_range})
        dataframe.to_csv(filename, index=False, sep=',', encoding="utf_8_sig")

    def ton_plot(self, y, stamping_counter):  # plot one new point on the ton curve
        self.save_ton_data(stamping_counter=stamping_counter-1)
        # print('ton ploting')
        output_length = 10
        # assert (self.num_channel == len(y))
        if self.num_channel == 0:
            return
        # stamping_counter = len(y[0])
        # total_ton_value = np.sum([ton_value[-1] for ton_value in y])
        # self.total_ton.append(total_ton_value)
        # print(total_ton_value)
        if stamping_counter >= output_length:
            self.total_ton = np.sum(
                [y[channel][stamping_counter - output_length:stamping_counter] for channel in self.ton_channels],
                axis=0)
        else:
            self.total_ton = np.sum([y[channel][:stamping_counter] for channel in self.ton_channels], axis=0)

        self.ton_curve.clear()
        self.ton_curve.addLegend(offset=(-1, 1))
        if stamping_counter < output_length:
            for channel in self.ton_channels:
                # 绘制各通道吨位曲线
                self.ton_curve.plot(pen=pg.mkPen(color=self.curve_color[channel], width=1.5),
                                    name="ch" + str(channel),
                                    symbolBrush=self.curve_color[channel]).setData(np.arange(10), np.concatenate(
                    (y[channel][:stamping_counter], np.zeros(output_length - stamping_counter))))
            # 绘制总吨位曲线
            # print((len(self.total_ton), output_length, stamping_counter, y))
            self.ton_curve.plot(pen=pg.mkPen(color=(130, 30, 130), width=1.5),
                                name="total",
                                symbolBrush=(130, 30, 130)).setData(np.arange(10), np.concatenate(
                (self.total_ton, np.zeros(output_length - stamping_counter))))
        else:
            # 绘制各通道吨位曲线
            for channel in self.ton_channels:
                self.ton_curve.plot(pen=pg.mkPen(color=self.curve_color[channel], width=1.5),
                                    name="ch" + str(channel),
                                    symbolBrush=self.curve_color[channel]).setData(
                    np.arange(stamping_counter - output_length, stamping_counter),
                    y[channel][stamping_counter - output_length: stamping_counter])
            # 绘制总吨位曲线
            self.ton_curve.plot(pen=pg.mkPen(color=(130, 30, 130), width=1.5),
                                name="total",
                                symbolBrush=(130, 30, 130)).setData(
                np.arange(stamping_counter - output_length, stamping_counter), self.total_ton)

        if stamping_counter > 0:
            for channel in self.ton_channels:
                if y[channel][stamping_counter - 1] > self.ton_range[channel][1] or y[channel][stamping_counter - 1] < \
                        self.ton_range[channel][0]:
                    self.signal_error.emit()
            if self.total_ton[-1] > self.ton_range[6][1] or self.total_ton[-1] < self.ton_range[6][0]:
                self.signal_error.emit()


class Runthread(QtCore.QThread):
    _signal = pyqtSignal(object, object, object, object)  # 原始数据信号
    _signal2 = pyqtSignal(int)  #
    _signal3 = pyqtSignal(str)  # 信号3 连接输出打印记录记录 textedit
    _signal4 = pyqtSignal(object, object, object, object, object, object, int)  # 用于绘制截取的信号绘制
    _signal5 = pyqtSignal(object, object, int, int, object)
    signal_ton = pyqtSignal(int, object, int)  # 用于传输数据给吨位计算
    signal_mode_change = pyqtSignal(int)  # 用于改变界面上的模式显示
    signal_calibration = pyqtSignal(int)  # 用于调整标定系数

    def __init__(self, obj=None, sdk=None, channel=None, curve=None, curve_win=None, save_mode=None,
                 mode_change_number=20, calibration_number=1, change_mode=True,
                 enable_predict=False, cut_length=10000, sampling_rate=64000, is_SG=False, fault_rate=0.1,
                 flag_calibration=False, output=None, model_name=None, fs=None):
        super(Runthread, self).__init__()
        self.flag = 1
        self.sdk = sdk
        ## 是否需要拉低复位
        # self.sdk.daq_io_set(HPS_DIO_PORT_2, HPS_DIO_LOW) # 拉低复位
        self.channel = str(channel)
        self.loop_data = []  # 用于每次更新数据  draw
        self.buffer_data = []  # 采集到的数据段 用于保存数据或预测模型
        self.index = 0  #
        self.curve = curve
        self.curve_win = curve_win
        self.save_mode = save_mode  # 模式选择0:只数据采集
        self.mode_change_number = mode_change_number
        self.calibration_number = calibration_number
        self.enable_predict = enable_predict
        self.cut_length = cut_length
        self.sampling_rate = sampling_rate
        self.is_SG = is_SG
        self.model_name = model_name
        self.fault_rate = fault_rate
        self.output = output
        self.fs = fs.replace(" ", "")
        self.flag_calibration = flag_calibration
        self.test_time = 0  # 用来进行吨位检测的数据保存及画图

        self._signal.connect(obj.updater)
        self._signal3.connect(obj.printf)
        self._signal4.connect(obj.updaterWin)
        self._signal5.connect(obj.runTime)

    def __del__(self):
        self.exiting = True
        self.wait()

    def stop(self):
        self.flag = 0

    def run(self):
        self.get_data_thread(int(self.channel), self.sdk, self.curve)

    # 触发刷新
    def data_emit(self, index, data):
        x = np.arange(index, len(data) + index)
        y = np.array(data)
        # print('data_emit working')
        if index == 0:
            clean = True
            self._signal.emit(self.curve, x, y, clean)  # 事件触发刷新
        else:
            clean = True
            self._signal.emit(self.curve, x, y, clean)  # 事件触发刷新

    def predict_main(self, data):
        # print(type(data))
        # print(np.shape(data))
        # print('predicting')
        t0 = time.time()
        # print(('max(data): ', max(data)))
        df_data = pd.DataFrame(data, columns=["channel_" + self.channel]).astype(np.float32)
        # print(('max(df_data): ', max(df_data.iloc[:, 0])))
        if self.save_mode >= 2 and self.enable_predict:
            # 调用python 模型预测结果
            predict_class = predict(channel=self.channel, data=df_data.iloc[:, 0], fault_rate=self.fault_rate,
                                    model_name=None)
            predict_class.main()
            signal = predict_class.cut_signal  # 信号1
            signal_upper = predict_class.upper
            signal_center = predict_class.center
            signal_down = predict_class.down
            # print((signal.shape, type(signal_center), signal_upper.shape))
            res = predict_class.res
            # print(('before', max(signal), max(signal_down), max(signal_upper)))
            # 暂时只有拉高
            if res == 1 and self.save_mode == 3:
                self.sdk.daq_io_set(HPS_DIO_PORT_2, HPS_DIO_HIGH)  # 拉高 # 反馈控制
            ## 触发更新信号包络曲线
            self._signal4.emit(self.curve_win, self.channel, signal, signal_upper, signal_center, signal_down, res)
            # print(type(df_data))
            # print(df_data.shape)
            self.save_collect_data(df_data, res=res)  # 保存原始数据

            self.save_handled_data(signal, signal_upper, signal_center, signal_down)
            # print(('after', max(signal), max(signal_down), max(signal_upper)))
            t1 = time.time()
            if predict_class.model["sample_num"] - predict_class.sample == 0:
                pass
                # print(predict_class.model["sample_num"] - predict_class.sample)
            self._signal5.emit(self.channel, round(t1 - t0, 6), predict_class.model["pass"],
                               predict_class.model["sample_num"] - predict_class.sample,
                               self.sdk)
        else:
            self.save_collect_data(df_data)  # 保存原始数据

    def ton_main(self, data, stamping_counter):
        # print(self.channel)
        self.signal_ton.emit(int(self.channel), data, stamping_counter)

    def save_collect_data(self, df_data, res=0):
        if self.save_mode <= 1:
            filename = "data_ch" + self.channel + "_" + self.fs + "_" + datetime.datetime.now().strftime(
                "%Y%m%d_%H%M_%S") + ".csv"
        else:  # 文件名后面加了判断的结果
            filename = "data_ch" + self.channel + "_" + self.fs + "_" + datetime.datetime.now().strftime(
                "%Y%m%d_%H%M_%S") + "_[" + str(res) + "]" + ".csv"
        df_data.to_csv("data\\" + filename, index=None, chunksize=2048)
        self._signal3.emit(filename)  # 发送弹幕

    def save_handled_data(self, data, data_up, data_center, data_down):
        filename = "data_ch" + self.channel + "_" + self.fs + "_" + datetime.datetime.now().strftime(
            "%Y%m%d_%H%M_%S")
        filename_data = filename + "data.csv"
        filename_up = filename + "up.csv"
        filename_center = filename + "center.csv"
        filename_down = filename + "down.csv"
        np.savetxt('handled\\' + filename_data, data, delimiter=',')
        np.savetxt("handled\\" + filename_up, data_up, delimiter=',')
        np.savetxt("handled\\" + filename_center, data_center, delimiter=',')
        np.savetxt("handled\\" + filename_down, data_down, delimiter=',')

    def get_data_thread(self, channel, sdk, curve):
        # print("running...")
        self._signal3.emit("running...")
        size = 100
        isw_file = DAQ_FALSE
        stamping_counter = 0
        display_count = 0
        while self.flag:
            ret, smp_data = sdk.daq_get_channel_data(channel, size, isw_file, CSV_FORMAT_DATA, 1000)
            if ret <= 0:
                time.sleep(0.01)  # 100ms
                # print(ret)
                continue
            read_size = ret
            # print('ret = ' + str(ret))
            self.loop_data = []  # 每次清空列表
            for i in range(0, read_size):
                if smp_data[i].sequence_num == END_FRAME_SEQ_NUM:
                    ##检验模型线程入口
                    ## 用于显示冲压信号
                    if self.flag_calibration:
                        continue
                    self.buffer_data = self.buffer_data[:self.cut_length]   # 截断信号，减少运算量，排除后段干扰
                    if self.is_SG:
                        for j in range(1, len(self.buffer_data)):
                            self.buffer_data[j] += self.buffer_data[j - 1]
                        for j in range(0, len(self.buffer_data)):
                            self.buffer_data[j] /= self.sampling_rate
                    data_emit = threading.Thread(target=self.data_emit, kwargs={
                        "index": self.index, "data": self.buffer_data})
                    data_emit.start()
                    ## 用于信号模型异常检测

                    stamping_counter += 1

                    self._signal3.emit("sequence_num =" + str(smp_data[i].sequence_num))
                    predict_thread = threading.Thread(target=self.predict_main, kwargs={
                        "data": self.buffer_data})
                    predict_thread.setDaemon(True)
                    predict_thread.start()

                    # 吨位检测
                    ton_thread = threading.Thread(target=self.ton_main,
                                                  kwargs={"data": self.buffer_data,
                                                          "stamping_counter": stamping_counter})
                    ton_thread.setDaemon(True)
                    ton_thread.start()

                    self.buffer_data = []
                    self.loop_data = []
                    self.index = 0
                    # display_count = 0
                    # print(stamping_counter)
                    if self.save_mode == 0 and stamping_counter == self.calibration_number:
                        # 吨位标定模式
                        # print('channel' + str(channel))

                        self.sdk.daq_io_set(HPS_DIO_PORT_2, HPS_DIO_HIGH)
                        self.signal_calibration.emit(channel)

                    # 纯采集模式，采集一定次数后进入异物监测+反馈控制模式
                    if stamping_counter >= self.mode_change_number and self.save_mode == 1:
                        self.save_mode = 3
                        self.signal_mode_change.emit(3)
                else:
                    data_list = [smp_data[i].vol_data[j] for j in range(0, smp_data[i].validnum)]  # 16个有效数据
                    if self.flag_calibration:
                        for index in range(len(data_list)):
                            data_list[index] /= self.sampling_rate
                            data_list[index] -= 0.009/20000
                    # print([smp_data[i].data[j] for j in range(0, smp_data[i].validnum)])  # 16个有效数据
                    self.buffer_data.extend(data_list)
                    self.loop_data.extend(data_list)
                    if self.flag_calibration:
                        display_count += 16
                        if len(self.buffer_data) == 16:
                            for index in range(1, 16):
                                self.buffer_data[index] += self.buffer_data[index - 1]
                        else:
                            for index in range(-16, 0):
                                self.buffer_data[index] += self.buffer_data[index - 1]
                    # print((self.flag_calibration, display_count))
                    if self.flag_calibration and display_count > 1600:
                        display_count = 0
                        self.buffer_data = self.buffer_data[-self.cut_length:]
                        # print(len(self.buffer_data))
                        # temp = self.buffer_data.copy()
                        # for index in range(len(temp)):
                        #     temp[index] /= self.sampling_rate
                        # for index in range(1, len(temp)):
                        #     temp[index] += temp[index - 1]
                        data_emit = threading.Thread(target=self.data_emit, kwargs={
                            "index": self.index, "data": self.buffer_data})
                        data_emit.start()

            else:
                # size * 16 100个循环  每次循环后触发刷新
                pass
                # if len(self.buffer_data) > 0:
                #     data_emit = threading.Thread(target=self.data_emit, kwargs={
                #         "index": self.index, "data": self.loop_data})
                #     data_emit.start()
                #     self.index += (size * 16)


class parentWindow(QMainWindow):
    # QTimer.thread()
    _signal = pyqtSignal(str)
    _connect_success = pyqtSignal()
    _disconnect = pyqtSignal()
    _line = pyqtSignal(int)
    _message_table = pyqtSignal()

    def __init__(self):
        QMainWindow.__init__(self)
        # 實例化頁面
        self.ui = Ui_MainWindow()
        # # 加載控件
        self.ui.setupUi(self)
        # 原始数据画布2*3布局
        self.raw = [pg.PlotWidget(background="#0d1f2d") for i in range(6)]
        self.win = [pg.PlotWidget(background="#0d1f2d") for i in range(6)]
        # 历史预测结果画布 (界面左下角)
        self.score = pg.PlotWidget(background="#d3eef3")
        # self.ton = pg.PlotWidget(background="#daddee")
        # 吨位检测画布
        self.ton = pg.PlotWidget(background="#0d1f2d")
        # self.ton_curve = pg.PlotWidget(background="#daddee")
        # self.ton.plot()

        self.VLayout_label_channel = QtWidgets.QVBoxLayout()
        self.ui.HLayout_ton.addLayout(self.VLayout_label_channel)
        new_label = QtWidgets.QLabel('通道')
        new_label.setAlignment(Qt.AlignCenter)
        self.VLayout_label_channel.addWidget(new_label)
        new_line = QtWidgets.QFrame()
        new_line.setFrameShape(QtWidgets.QFrame.VLine)
        new_line.setFrameShadow(QtWidgets.QFrame.Sunken)
        self.ui.HLayout_ton.addWidget(new_line)
        self.VLayout_ton_real = QtWidgets.QVBoxLayout()
        self.ui.HLayout_ton.addLayout(self.VLayout_ton_real)
        new_label = QtWidgets.QLabel('标定吨位')
        new_label.setAlignment(Qt.AlignCenter)
        self.VLayout_ton_real.addWidget(new_label)
        new_line = QtWidgets.QFrame()
        new_line.setFrameShape(QtWidgets.QFrame.VLine)
        new_line.setFrameShadow(QtWidgets.QFrame.Sunken)
        self.ui.HLayout_ton.addWidget(new_line)
        # self.VLayout_spinBox_ton_real.setAlignment(Qt.AlignCenter)
        self.VLayout_ton_calibration = QtWidgets.QVBoxLayout()
        self.ui.HLayout_ton.addLayout(self.VLayout_ton_calibration)
        new_label = QtWidgets.QLabel('标定系数')
        new_label.setAlignment(Qt.AlignCenter)
        self.VLayout_ton_calibration.addWidget(new_label)
        new_line = QtWidgets.QFrame()
        new_line.setFrameShape(QtWidgets.QFrame.VLine)
        new_line.setFrameShadow(QtWidgets.QFrame.Sunken)
        self.ui.HLayout_ton.addWidget(new_line)
        self.VLayout_ton_calculated = QtWidgets.QVBoxLayout()
        self.ui.HLayout_ton.addLayout(self.VLayout_ton_calculated)
        new_label = QtWidgets.QLabel('实际吨位')
        new_label.setAlignment(Qt.AlignCenter)
        self.VLayout_ton_calibration.addWidget(new_label)
        self.VLayout_ton_calculated.addWidget(new_label)
        new_line = QtWidgets.QFrame()
        new_line.setFrameShape(QtWidgets.QFrame.VLine)
        new_line.setFrameShadow(QtWidgets.QFrame.Sunken)
        self.ui.HLayout_ton.addWidget(new_line)
        self.VLayout_ton_range = QtWidgets.QVBoxLayout()
        self.ui.HLayout_ton.addLayout(self.VLayout_ton_range)
        new_label = QtWidgets.QLabel('报警吨位')
        new_label.setAlignment(Qt.AlignCenter)
        self.VLayout_ton_calibration.addWidget(new_label)
        self.VLayout_ton_range.addWidget(new_label)
        new_line = QtWidgets.QFrame()
        new_line.setFrameShape(QtWidgets.QFrame.VLine)
        new_line.setFrameShadow(QtWidgets.QFrame.Sunken)
        self.ui.HLayout_ton.addWidget(new_line)

        self.label_ton_channel = []
        self.spinBox_ton_real = []
        self.spinBox_ton_calibration = []
        self.label_ton_calculated = []
        self.spinBox_ton_range = []
        channel_label_list = ['ch' + str(index) for index in range(6)]
        for channel_label in channel_label_list:
            new_label = QtWidgets.QLabel(channel_label)
            new_label.setAlignment(Qt.AlignCenter)
            self.label_ton_channel.append(new_label)
            self.VLayout_label_channel.addWidget(new_label)

            # for ton_layout in self.ton_layouts:
            new_spinBox_ton_real = QtWidgets.QDoubleSpinBox()
            new_spinBox_ton_real.setAlignment(QtCore.Qt.AlignCenter)
            new_spinBox_ton_real.setMaximum(9999)
            self.spinBox_ton_real.append(new_spinBox_ton_real)
            new_spinBox_ton_calibration = QtWidgets.QDoubleSpinBox()
            new_spinBox_ton_calibration.setValue(1)
            new_spinBox_ton_calibration.setMaximum(999999)
            new_spinBox_ton_calibration.setDecimals(3)
            new_spinBox_ton_calibration.setAlignment(QtCore.Qt.AlignCenter)
            self.spinBox_ton_calibration.append(new_spinBox_ton_calibration)
            new_label_ton_calculated = QtWidgets.QLabel()
            new_label_ton_calculated.setAlignment(QtCore.Qt.AlignCenter)
            self.label_ton_calculated.append(new_label_ton_calculated)
            new_label = QtWidgets.QLabel('-')
            new_spinBox_ton_range = [QtWidgets.QSpinBox(), QtWidgets.QSpinBox(), new_label]
            new_spinBox_ton_range[0].setAlignment(QtCore.Qt.AlignRight)
            new_spinBox_ton_range[1].setAlignment(QtCore.Qt.AlignLeft)
            self.spinBox_ton_range.append(new_spinBox_ton_range)

            self.VLayout_ton_real.addWidget(new_spinBox_ton_real)
            self.VLayout_ton_calibration.addWidget(new_spinBox_ton_calibration)
            self.VLayout_ton_calculated.addWidget(new_label_ton_calculated)
            new_HLayout = QtWidgets.QHBoxLayout()
            new_HLayout.addWidget(new_spinBox_ton_range[0])
            new_HLayout.addWidget(new_label)
            new_label.setAlignment(QtCore.Qt.AlignCenter)
            new_label.setMaximumWidth(30)
            new_HLayout.addWidget(new_spinBox_ton_range[1])
            new_spinBox_ton_range[0].setMaximum(9999)
            new_spinBox_ton_range[1].setMaximum(9999)
            new_spinBox_ton_range[1].setValue(9999)
            self.VLayout_ton_range.addLayout(new_HLayout)

        new_label = QtWidgets.QLabel('total')
        new_label.setAlignment(Qt.AlignCenter)
        self.label_ton_channel.append(new_label)
        self.VLayout_label_channel.addWidget(new_label)

        # for ton_layout in self.ton_layouts:
        new_spinBox_ton_real = QtWidgets.QPushButton()
        new_spinBox_ton_real.setText('自动校准')
        # new_spinBox_ton_real.setAlignment(QtCore.Qt.AlignCenter)
        self.spinBox_ton_real.append(new_spinBox_ton_real)
        new_spinBox_ton_calibration = QtWidgets.QPushButton()
        new_spinBox_ton_calibration.setText('手动校准')
        # new_spinBox_ton_calibration.setAlignment(QtCore.Qt.AlignCenter)
        self.spinBox_ton_calibration.append(new_spinBox_ton_calibration)
        new_label_ton_calculated = QtWidgets.QLabel()
        new_label_ton_calculated.setAlignment(QtCore.Qt.AlignCenter)
        self.label_ton_calculated.append(new_label_ton_calculated)
        new_label = QtWidgets.QLabel('-')
        new_spinBox_ton_range = [QtWidgets.QSpinBox(), QtWidgets.QSpinBox(), new_label]
        new_spinBox_ton_range[0].setAlignment(QtCore.Qt.AlignRight)
        new_spinBox_ton_range[1].setAlignment(QtCore.Qt.AlignLeft)
        self.spinBox_ton_range.append(new_spinBox_ton_range)

        self.VLayout_ton_real.addWidget(new_spinBox_ton_real)
        self.VLayout_ton_calibration.addWidget(new_spinBox_ton_calibration)
        self.VLayout_ton_calculated.addWidget(new_label_ton_calculated)
        new_HLayout = QtWidgets.QHBoxLayout()
        new_HLayout.addWidget(new_spinBox_ton_range[0])
        new_HLayout.addWidget(new_label)
        new_label.setAlignment(QtCore.Qt.AlignCenter)
        new_label.setMaximumWidth(30)
        new_HLayout.addWidget(new_spinBox_ton_range[1])
        new_spinBox_ton_range[0].setMaximum(9999)
        new_spinBox_ton_range[1].setMaximum(9999)
        new_spinBox_ton_range[1].setValue(9999)
        self.VLayout_ton_range.addLayout(new_HLayout)

        self.spinBox_ton_real[-1].setEnabled(False)
        self.spinBox_ton_calibration[-1].setEnabled(False)
        self.spinBox_ton_range = np.array(self.spinBox_ton_range)

        for spinBox in self.spinBox_ton_real + self.spinBox_ton_calibration:
            spinBox.setEnabled(False)
        self.tonLayout = QVBoxLayout()
        # self.tonLayout1 = QVBoxLayout()
        # self.ton.showAxis("left", False)
        # self.ton.showAxis("bottom", False)
        self.ton.showGrid(x=True, y=True)
        # self.ton_curve.showGrid(x=True, y=True)
        self.tonLayout.addWidget(self.ton)
        # self.tonLayout1.addWidget(self.ton_curve)
        self.ui.graphicsView_4.setLayout(self.tonLayout)
        # self.ui.graphicsView_7.setLayout(self.tonLayout1)
        # channel
        # self.get_channel()
        self.channel = ["channel0", "channel1", "channel2", "channel3", "channel4", "channel5"]
        # self.ton_channels = self.channel[:4]
        #
        # self.ton_class = ton_class(ton_curve=self.ton, ton_curve_win=None, ton_channels=self.ton_channels, fs=None)
        # 模拟数据
        # self.data=pd.read_csv(r"C:\Users\liu\Desktop\CanisPro-DAQ\data_0910\data\data-chan2_20200910_1915_14.csv")

        pltItem = [pw.getPlotItem() for pw in self.raw]
        bottom_axis = [pltItem.getAxis("bottom") for pltItem in pltItem]
        [bottom.setLabel(channel, ) for channel, bottom in
         zip(self.channel, bottom_axis)]

        pltItem = [pw.getPlotItem() for pw in self.win]
        bottom_axis = [pltItem.getAxis("bottom") for pltItem in pltItem]
        [bottom.setLabel(channel, ) for channel, bottom in
         zip(self.channel, bottom_axis)]

        # self.ton_class = ton_class(ton_curve=self.ton, ton_curve_win=self.tonLayout)

        # self._pen() # 画画的baby
        self._layoutSetting()  # 布局画布设定
        self._init()  # 初始化信号与槽的连接
        self._loadModel()  # 加载离线模型

        self.res = 0
        self.result_list = []  # 用于储存历史预测结果
        self.show_result = 0

    def _init(self):
        self.pen_red = pg.mkPen(color=(255, 0, 0), width=2)
        self.pen_green = pg.mkPen(color=(0, 255, 0), width=2)
        self.pen_white = pg.mkPen(color=(250, 250, 250), width=2)
        self.pen_black = pg.mkPen(color=(10, 100, 100), width=2)
        self.pen_yellow = pg.mkPen(color=(255, 255, 10), width=2)
        self.pen_yellow_2 = pg.mkPen(color=(255, 200, 150), width=2)
        self.model_display = [
            (self.ui.box_model1_sample, self.ui.box_model1_filterfre, self.ui.box_model1_sigleft,
             self.ui.box_model1_sigright, self.ui.box_model1_rate, self.ui.label_score1, self.ui.label_time_1),
            (self.ui.box_model2_sample, self.ui.box_model2_filterfre, self.ui.box_model2_sigleft,
             self.ui.box_model2_sigright, self.ui.box_model2_rate, self.ui.label_score2, self.ui.label_time_2),
            (self.ui.box_model3_sample, self.ui.box_model3_filterfre, self.ui.box_model3_sigleft,
             self.ui.box_model3_sigright, self.ui.box_model3_rate, self.ui.label_score3, self.ui.label_time_3),
            (self.ui.box_model4_sample, self.ui.box_model4_filterfre, self.ui.box_model4_sigleft,
             self.ui.box_model4_sigright, self.ui.box_model4_rate, self.ui.label_score4, self.ui.label_time_4),
            (self.ui.box_model5_sample, self.ui.box_model5_filterfre, self.ui.box_model5_sigleft,
             self.ui.box_model5_sigright, self.ui.box_model5_rate, self.ui.label_score5, self.ui.label_time_5),
            (self.ui.box_model6_sample, self.ui.box_model6_filterfre, self.ui.box_model6_sigleft,
             self.ui.box_model6_sigright, self.ui.box_model6_rate, self.ui.label_score6, self.ui.label_time_6),
        ]
        # channel
        # self.ui.checkBox_ton_all.stateChanged.connect(lambda: self.stateChannel(self.ui.checkBox_ton_all.checkState()))
        # connect
        self.ui.btn_connect.clicked.connect(self.connectDaq)
        # disconnect
        self.ui.btn_disconnect.clicked.connect(self.disconnectDaq)
        # reset
        self.ui.btn_reset.clicked.connect(lambda: self.sdk.daq_io_set(HPS_DIO_PORT_2, HPS_DIO_LOW))
        self.ui.btn_reset.setEnabled(False)
        # model
        self.ui.btn_model1_change.clicked.connect(
            lambda: self.changeModel(self.model[0], self.model_display[0], "save_model_0.pk"))
        self.ui.btn_model2_change.clicked.connect(
            lambda: self.changeModel(self.model[1], self.model_display[1], "save_model_1.pk"))
        self.ui.btn_model3_change.clicked.connect(
            lambda: self.changeModel(self.model[2], self.model_display[2], "save_model_2.pk"))
        self.ui.btn_model4_change.clicked.connect(
            lambda: self.changeModel(self.model[3], self.model_display[3], "save_model_3.pk"))
        self.ui.btn_model5_change.clicked.connect(
            lambda: self.changeModel(self.model[4], self.model_display[4], "save_model_4.pk"))
        self.ui.btn_model6_change.clicked.connect(
            lambda: self.changeModel(self.model[5], self.model_display[5], "save_model_5.pk"))
        self.ui.btn_change_all.clicked.connect(lambda: self.change_all())

        self.ui.btn_model1_retraining.clicked.connect(
            lambda: self.modelRetraining(self.model[0], self.model_display[0], "save_model_0.pk"))
        self.ui.btn_model2_retraining.clicked.connect(
            lambda: self.modelRetraining(self.model[1], self.model_display[1], "save_model_1.pk"))
        self.ui.btn_model3_retraining.clicked.connect(
            lambda: self.modelRetraining(self.model[2], self.model_display[2], "save_model_2.pk"))
        self.ui.btn_model4_retraining.clicked.connect(
            lambda: self.modelRetraining(self.model[3], self.model_display[3], "save_model_3.pk"))
        self.ui.btn_model5_retraining.clicked.connect(
            lambda: self.modelRetraining(self.model[4], self.model_display[4], "save_model_4.pk"))
        self.ui.btn_model6_retraining.clicked.connect(
            lambda: self.modelRetraining(self.model[5], self.model_display[5], "save_model_5.pk"))
        self.ui.btn_retrain_all.clicked.connect(lambda: self.retrain_all())

        self.ui.btn_model1_more.clicked.connect(lambda: self.model_more(self.model[0], "模型1的历史预测结果"))
        self.ui.btn_model2_more.clicked.connect(lambda: self.model_more(self.model[1], "模型2的历史预测结果"))
        self.ui.btn_model3_more.clicked.connect(lambda: self.model_more(self.model[2], "模型3的历史预测结果"))
        self.ui.btn_model4_more.clicked.connect(lambda: self.model_more(self.model[3], "模型4的历史预测结果"))
        self.ui.btn_model5_more.clicked.connect(lambda: self.model_more(self.model[4], "模型5的历史预测结果"))
        self.ui.btn_model6_more.clicked.connect(lambda: self.model_more(self.model[5], "模型6的历史预测结果"))

        self.timer = None
        self.ui.btn_cleardata.clicked.connect(lambda: self.clear_data(tip=True))
        self.ui.rb_999.toggled.connect(lambda: self.toggleddd(self.ui.rb_999, 999999))
        self.ui.rb_3.toggled.connect(lambda: self.toggleddd(self.ui.rb_3, 3))
        self.ui.rb_7.toggled.connect(lambda: self.toggleddd(self.ui.rb_7, 7))
        self.ui.rb_15.toggled.connect(lambda: self.toggleddd(self.ui.rb_15, 15))
        self.ui.rb_30.toggled.connect(lambda: self.toggleddd(self.ui.rb_30, 30))
        self.ui.checkBox_ton_all.stateChanged.connect(self.checkTonAllChannel)  # 設置全選按鈕的觸發鏈接
        self.ui.checkBox_abnormal_all.stateChanged.connect(self.checkAbnormalAllChannel)
        self.ui.checkBox_SG_all.stateChanged.connect(self.checkSGALLChannel)

        self.ui.checkBox_ton_ch0.stateChanged.connect(
            lambda state: self.ui.checkBox_abnormal_ch0.setChecked(False) if state else None)
        self.ui.checkBox_abnormal_ch0.stateChanged.connect(
            lambda state: self.ui.checkBox_ton_ch0.setChecked(False) if state else None)
        self.ui.checkBox_ton_ch1.stateChanged.connect(
            lambda state: self.ui.checkBox_abnormal_ch1.setChecked(False) if state else None)
        self.ui.checkBox_abnormal_ch1.stateChanged.connect(
            lambda state: self.ui.checkBox_ton_ch1.setChecked(False) if state else None)
        self.ui.checkBox_ton_ch2.stateChanged.connect(
            lambda state: self.ui.checkBox_abnormal_ch2.setChecked(False) if state else None)
        self.ui.checkBox_abnormal_ch2.stateChanged.connect(
            lambda state: self.ui.checkBox_ton_ch2.setChecked(False) if state else None)
        self.ui.checkBox_ton_ch3.stateChanged.connect(
            lambda state: self.ui.checkBox_abnormal_ch3.setChecked(False) if state else None)
        self.ui.checkBox_abnormal_ch3.stateChanged.connect(
            lambda state: self.ui.checkBox_ton_ch3.setChecked(False) if state else None)
        self.ui.checkBox_ton_ch4.stateChanged.connect(
            lambda state: self.ui.checkBox_abnormal_ch4.setChecked(False) if state else None)
        self.ui.checkBox_abnormal_ch4.stateChanged.connect(
            lambda state: self.ui.checkBox_ton_ch4.setChecked(False) if state else None)
        self.ui.checkBox_ton_ch5.stateChanged.connect(
            lambda state: self.ui.checkBox_abnormal_ch5.setChecked(False) if state else None)
        self.ui.checkBox_abnormal_ch5.stateChanged.connect(
            lambda state: self.ui.checkBox_ton_ch5.setChecked(False) if state else None)

        self._signal.connect(self.printf)
        self._connect_success.connect(self.connect_sucessful)
        self._disconnect.connect(self.disconnect_successful)
        self._message_table.connect(self.table_insert)
        self._line.connect(self._line_bar)
        # self.score_plot2([1,1,1,-1,-1,1,-1,1,-1,1,-1, 1,1,1,1,1,1,1,])
        self.score_plot2([1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, ])

        # self.score_plot2([-1])

        # self._plot()
        # self.score_plot2([-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1])
        # vLine = pg.InfiniteLine(angle=90, movable=False, )
        # hLine = pg.InfiniteLine(angle=0, movable=False, )
        # self.ton.addItem(vLine, ignoreBounds=True)
        # self.ton.addItem(hLine, ignoreBounds=True)
        self.ton.addLegend(offset=(-1, 1))
        # self.ton_curve.addLegend(offset=(0, 0))
        # self.ton.plot(pen=pg.mkPen(color=(255, 0, 0), width=1.5), name="ch0", ).setData([1,2,3,4,5,6,7,8,9,10], self.get_random_value())
        # self.ton.plot(pen=pg.mkPen(color=(255, 200, 150), width=1.5), name="ch1", ).setData([1,2,3,4,5,6,7,8,9,10], self.get_random_value())
        # self.ton.plot(pen=pg.mkPen(color=(10, 100, 100), width=1.5), name="ch2", ).setData([1,2,3,4,5,6,7,8,9,10], self.get_random_value())
        # self.ton.plot(pen=pg.mkPen(color=(250, 250, 250), width=1.5), name="ch3",).setData([1,2,3,4,5,6,7,8,9,10], self.get_random_value())
        # self.ton.plot(pen=pg.mkPen(color=(0, 255, 0), width=1.5), name="ch4",).setData([1,2,3,4,5,6,7,8,9,10], self.get_random_value())

        # 吨位检测曲线
        # self.ton.plot(pen=pg.mkPen(color=(255, 0, 0), width=1.5), name="ch5", symbolBrush=(255, 0, 0)).setData(
        #     np.arange(10), self.get_random_value(random=True) * 250)
        # self.ton_curve.plot(pen=pg.mkPen(color=(0, 0, 0), width=2), name="Reference curve").setData(np.arange(15),
        #                                                                                             self.get_random_value(
        #                                                                                                 15, add=-.2))
        # self.ton_curve.plot(pen=pg.mkPen(color=(48, 68, 185), width=2), name="Current curve").setData(np.arange(15),
        #                                                                                               self.get_random_value(
        #                                                                                                   15))

    def get_channel(self):
        self.ton_channel_list = [
            self.ui.checkBox_ton_ch0.isChecked(),
            self.ui.checkBox_ton_ch1.isChecked(),
            self.ui.checkBox_ton_ch2.isChecked(),
            self.ui.checkBox_ton_ch3.isChecked(),
            self.ui.checkBox_ton_ch4.isChecked(),
            self.ui.checkBox_ton_ch5.isChecked(),
        ]
        self.channel_list = self.ton_channel_list

    def ton_error(self):
        # self.disconnectDaq()
        if self.ui.radio_mode3.isChecked():
            self.sdk.daq_io_set(HPS_DIO_PORT_2, HPS_DIO_HIGH)

    def checkTonAllChannel(self, state):
        # if state == QtCore.Qt.Checked:
        self.ui.checkBox_ton_ch0.setChecked(state)
        self.ui.checkBox_ton_ch1.setChecked(state)
        self.ui.checkBox_ton_ch2.setChecked(state)
        self.ui.checkBox_ton_ch3.setChecked(state)
        self.ui.checkBox_ton_ch4.setChecked(state)
        self.ui.checkBox_ton_ch5.setChecked(state)

    def checkAbnormalAllChannel(self, state):
        # if state == QtCore.Qt.Checked:
        self.ui.checkBox_abnormal_ch0.setChecked(state)
        self.ui.checkBox_abnormal_ch1.setChecked(state)
        self.ui.checkBox_abnormal_ch2.setChecked(state)
        self.ui.checkBox_abnormal_ch3.setChecked(state)
        self.ui.checkBox_abnormal_ch4.setChecked(state)
        self.ui.checkBox_abnormal_ch5.setChecked(state)

    def checkSGALLChannel(self, state):
        self.ui.checkBox_SG_ch0.setChecked(state)
        self.ui.checkBox_SG_ch1.setChecked(state)
        self.ui.checkBox_SG_ch2.setChecked(state)
        self.ui.checkBox_SG_ch3.setChecked(state)
        self.ui.checkBox_SG_ch4.setChecked(state)
        self.ui.checkBox_SG_ch5.setChecked(state)

    def checkTonChannel(self, checkBox, state):
        checkBox.setChecked(~state)

    def toggleddd(self, target, Time):
        if target.isChecked():
            if Time > 100:
                self.timer = None
            else:
                if self.timer is None:
                    self.create_timer(Time)
                else:
                    self.timer.stop()
                    self.create_timer(Time)

    def create_timer(self, Time):
        self.timer = QTimer()
        self.timer.timeout.connect(lambda: self.clear_data(tip=False))
        self.timer.start(Time * 1000 * 60 * 60 * 24)  # day
        # self.timer.start(Time * 1000)  # s

    def clear_data(self, tip=True):
        try:
            if not tip:
                self.delete()
                return None
            reply = QMessageBox.information(self, 'Tip', '数据清空将不可撤销!!!\n请谨慎操作！！确认按<ok>',
                                            QMessageBox.Ok | QMessageBox.Close)
            if reply == 1024:
                self.delete()

        except Exception as e:
            print(e)

    def delete(self):
        try:
            shutil.rmtree("data")
            os.mkdir("data")
        except:
            time.sleep(0.5)  ##延时一会在创建一个新目录
            os.mkdir("data")
        try:
            shutil.rmtree("handled")
            os.mkdir("handled")
        except:
            time.sleep(0.5)  ##延时一会在创建一个新目录
            os.mkdir("handled")

    def get_random_value(self, num=10, random=False, add=0.):
        value = np.hamming(num)
        if random:
            r = np.random.random(size=num)
            value = value + r
        return value + add
        # return value.tolist()

    def _line_bar(self, p):
        self.ui.progressBar.setValue(p)
        self.printf(str(p) + "%")

    # insert,只是简单插入一个固定数据
    def table_insert(self):
        self.ui.tableWidget.horizontalHeader().setDefaultSectionSize(10)
        # self.ui.tableWidget.horizontalHeader().setColumnWidth(0, 50)
        row = self.ui.tableWidget.rowCount()
        self.ui.tableWidget.insertRow(row)

        item_timestamp = QTableWidgetItem("2020-10-10 08:00:00")
        item_channel = QTableWidgetItem("1")
        item_filename = QTableWidgetItem("test.csv")
        item_status = QTableWidgetItem("Abnormal")
        item_step = QTableWidgetItem("收集样本")
        item_object = QTableWidgetItem("金属片")
        item_message = QTableWidgetItem("Play Game")

        self.ui.tableWidget.setItem(row, 0, item_timestamp)
        self.ui.tableWidget.setItem(row, 1, item_channel)
        self.ui.tableWidget.setItem(row, 2, item_filename)
        self.ui.tableWidget.setItem(row, 3, item_status)
        self.ui.tableWidget.setItem(row, 4, item_step)
        self.ui.tableWidget.setItem(row, 5, item_object)
        self.ui.tableWidget.setItem(row, 6, item_message)

    def model_more(self, model, title):
        win = pg.plot(background="#0d1f2d", title=title)
        win.showGrid(x=True, y=True)
        res = model["res"][-100:]
        x1 = np.argwhere(res == 1)
        x2 = np.argwhere(res == 0)

        y = np.ones(res.shape[0]) - 0.5
        # create horizontal list
        x = np.arange(100)
        np.random.shuffle(x)
        bg1 = pg.BarGraphItem(x=x1, height=y[x1], width=1, brush='r')
        bg2 = pg.BarGraphItem(x=x2, height=y[x2], width=1, brush='g')
        win.addItem(bg1)
        win.addItem(bg2)

    def modelRetraining(self, model, model_display, model_name, ask=True):
        if ask:
            reply = QMessageBox.information(self, 'Tip', '重新训练Model后, 之前的样本数据将会清空\n请谨慎操作！！确认按<ok>',
                                            QMessageBox.Ok | QMessageBox.Close)
        else:
            reply = 1024
        if reply == 1024:
            model["sample_num"] = 0
            model["sample_data"] = None
            model["upper_lim"] = None
            model["down_lim"] = None
            model["mean"] = None
            model["res"] = np.array([])
            model["pass"] = 0
            with open("model//" + model_name, "wb") as target:
                dump(model, target)
            model_display[5].setText("0.0%")
            model_display[5].setToolTip("pass: 0\ntotal: 0")
            self._signal.emit(model_name + " Retraining oK!")

    def changeModel(self, model, model_display, model_name, need_reply=True):
        model["sample"] = model_display[0].value()
        model["filter_fre"] = model_display[1].value()
        model["sig_left"] = model_display[2].value()
        model["sig_right"] = model_display[3].value()
        model["fault_rate"] = model_display[4].value()
        with open("model//" + model_name, "wb") as target:
            dump(model, target)

        if need_reply:
            reply = QMessageBox.information(self, 'Tip', model_name + ' saved successfully!', QMessageBox.Close)
        self._signal.emit(model_name + " Change Sucessfully")

    def change_all(self):
        for channel in range(6):
            self.changeModel(self.model[channel], self.model_display[channel], 'save_model_%d.pk' % channel, need_reply=False)

    def retrain_all(self):
        for channel in range(6):
            self.modelRetraining(self.model[channel], self.model_display[channel], 'save_model_%d.pk' % channel, ask=False)

    def _loadModel(self):
        try:
            os.mkdir("model")
            os.mkdir("data")
            os.mkdir("handled")
            os.mkdir("tondata")
        except:
            pass
        file_list = os.listdir("model//")
        model_pk_file = [f for f in file_list if ".pk" in f]
        self.model = []

        for i, display in zip(range(6), self.model_display):
            model_name = "save_model_" + str(i) + ".pk"
            model = None
            if model_name not in model_pk_file:
                init_model = {
                    "upper_lim": None,
                    "down_lim": None,
                    "mean": None,
                    "sigma": 3,
                    "fault_rate": 0.1,
                    "fs": 64000,
                    "sample": 10,
                    "sample_num": 0,
                    "sample_data": None,
                    "sig_left": 1000,
                    "sig_right": 800,
                    "filter_fre": 100,
                    "res": np.array([]),
                    "pass": 0,
                    "name": model_name,
                    "model_version": "0.01"
                }
                with open("model//" + model_name, "wb") as target:
                    dump(init_model, target)
                model = init_model
                self.model.append(model)
            else:
                with open("model//" + model_name, "rb") as target:
                    model = load(target)
                self.model.append(model)
            display[0].setValue(model.get("sample"))
            display[1].setValue(model.get("filter_fre"))
            display[2].setValue(model.get("sig_left"))
            display[3].setValue(model.get("sig_right"))
            display[4].setValue(model.get("fault_rate"))

    def printf(self, text):
        self.ui.textEdit.append(">> " + text)

    def _collect(self, channel_list=None):
        self.sdk = DaqSDK()
        # self.ui.progressBar.setValue(50)  # 70%
        self._line.emit(50)
        rate_index = self.ui.edit_connect_rate.currentIndex()
        rate_str = self.ui.edit_connect_rate.currentText()
        rate = int(rate_str[: -1]) * 1000
        mode = self.ui.edit_connect_input.currentText()
        flag_calibration = mode == '持续采集'
        ret = run_core._init(self.sdk, self._signal, rate_index)  # 3 === 64K
        self.Runthread_list = []
        if ret:
            ret = self.sdk.daq_io_set(HPS_DIO_PORT_2, HPS_DIO_LOW)  # 拉低
            self._signal.emit("daq connect successfully")
            # self.ui.progressBar.setValue(70)  # 70%
            self._line.emit(70)
            for channel, enable in enumerate(channel_list):
                if enable:
                    print('channel = ' + str(channel))
                    self.Runthread = Runthread(obj=self,
                                               sdk=self.sdk,
                                               channel=channel,
                                               curve=self.raw[channel],
                                               curve_win=self.win[channel],
                                               save_mode=self.current_mode_index,
                                               mode_change_number=self.ui.spinBox_mode_change_number.value(),
                                               calibration_number=self.ui.spinBox_calibration_time.value(),
                                               change_mode=self.ui.checkBox_mode_change.isChecked(),
                                               enable_predict=self.channel_abnormal_list[channel],
                                               cut_length=self.ui.spinBox_cutlength.value(),
                                               sampling_rate=rate,
                                               is_SG=self.channel_SG_list[channel],
                                               fault_rate=self.model[channel]["fault_rate"],
                                               flag_calibration=flag_calibration,
                                               output="",
                                               model_name="",
                                               fs=rate_str,
                                               )
                    self.Runthread.signal_ton.connect(self.ton_class.receive_data)
                    # self.ton_class.signal_error.connect(self.ton_error)
                    self.Runthread.signal_calibration.connect(self.ton_class.change_calibration)
                    self.Runthread.signal_mode_change.connect(self.change_mode)
                    self.Runthread.start()
                    self.Runthread_list.append(self.Runthread)
            # self.ui.btn_connect.setEnabled(False)
            # self.ui.btn_reset.setEnabled(True)
            # self.ui.box_ton_min.setEnabled(False)
            # self.ui.box_ton_max.setEnabled(False)
            # self.ton_range = [self.ui.box_ton_min.value(), self.ui.box_ton_max.value()]
            # self.ton_class.set_ton_range(self.spinBox_ton_range)
            # self.ui.progressBar.setValue(100)  # 70%
            self._line.emit(100)
            # self.timer2.start(100)  #实时刷新页面
            # self.connect_sucessful()
            self._connect_success.emit()
            while self._flag:
                run_core.checkStatus(self.sdk)
            for thead in self.Runthread_list:
                thead.stop()
            self.sdk.daq_disconnect()

        # self.disconnect_successful()
        self._disconnect.emit()

    def disconnectDaq(self):
        self._flag = 0
        # self._message_table.emit()

    def disconnect_successful(self):
        for i in range(101):
            self.ui.progressBar.setValue(100 - i)
            time.sleep(.001)
        self.ui.groupBox_5.setEnabled(True)
        self.ui.groupBox_4.setEnabled(True)
        # self.ui.groupBox_3.setEnabled(True)
        self.ui.btn_connect.setEnabled(True)
        self.ui.btn_reset.setEnabled(False)
        for spinBox in self.spinBox_ton_real + self.spinBox_ton_calibration:
            spinBox.setEnabled(False)
        for HBox in self.spinBox_ton_range:
            for spinBox in HBox:
                spinBox.setEnabled(True)
        self.ui.checkBox_mode_change.setEnabled(True)
        self.spinBox_ton_real[-1].clicked.disconnect(self.ton_class.calculate_calibration)
        self.spinBox_ton_calibration[-1].clicked.disconnect(self.ton_class.update_ton_calibration_value)
        # self.ui.box_ton_min.setEnabled(True)
        # self.ui.box_ton_max.setEnabled(True)
        self._signal.emit("daq disconnect!")

    def connectDaq(self):
        print('connect')
        # 判定channel
        self.channel_ton_list = [
            self.ui.checkBox_ton_ch0.isChecked(),
            self.ui.checkBox_ton_ch1.isChecked(),
            self.ui.checkBox_ton_ch2.isChecked(),
            self.ui.checkBox_ton_ch3.isChecked(),
            self.ui.checkBox_ton_ch4.isChecked(),
            self.ui.checkBox_ton_ch5.isChecked(),
        ]
        self.channel_SG_list = [
            self.ui.checkBox_SG_ch0.isChecked(),
            self.ui.checkBox_SG_ch1.isChecked(),
            self.ui.checkBox_SG_ch2.isChecked(),
            self.ui.checkBox_SG_ch3.isChecked(),
            self.ui.checkBox_SG_ch4.isChecked(),
            self.ui.checkBox_SG_ch5.isChecked(),
        ]
        self.ton_channels = []
        for channel_index, enable in zip(range(len(self.channel_ton_list)), self.channel_ton_list):
            if enable:
                self.ton_channels.append(channel_index)

        # 这一块需要移到连接sdk之后
        for HBox in self.spinBox_ton_range:
            for spinBox in HBox:
                spinBox.setEnabled(False)
        self.ton_class = ton_class(ton_curve=self.ton, ton_curve_win=None, ton_channels=self.ton_channels, SG_list=self.channel_SG_list, fs=None)
        self.ton_class.set_output_label(self.label_ton_calculated)
        self.ton_class.set_ton_range(self.spinBox_ton_range)
        self.ton_class.set_calibration_time(self.ui.spinBox_calibration_time.value())
        self.ton_class.set_spinBox_ton_real(self.spinBox_ton_real)
        self.ton_class.set_spinBox_ton_calibration(self.spinBox_ton_calibration)
        self.ton_class.update_ton_calibration_value()
        self.ton_class.signal_error.connect(self.ton_error)
        self.spinBox_ton_calibration[-1].clicked.connect(self.ton_class.update_ton_calibration_value)
        self.spinBox_ton_real[-1].clicked.connect(self.ton_class.calculate_calibration)
        self.ui.btn_calibration_reset.clicked.connect(self.ton_class.reset_calibration)
        # for channel in self.ton_channels:
        #     self.ton_class.receive_data(channel, [0, channel, 0] + [0 for i in range(2000)], 1)
        # self.ton_class.ton_plot([[i + 3] for i in self.ton_channels], 1)

        self.channel_abnormal_list = [
            self.ui.checkBox_abnormal_ch0.isChecked(),
            self.ui.checkBox_abnormal_ch1.isChecked(),
            self.ui.checkBox_abnormal_ch2.isChecked(),
            self.ui.checkBox_abnormal_ch3.isChecked(),
            self.ui.checkBox_abnormal_ch4.isChecked(),
            self.ui.checkBox_abnormal_ch5.isChecked(),
        ]
        print(self.channel_ton_list)
        self.channel_list = np.logical_or(self.channel_ton_list, self.channel_abnormal_list)
        # print(self.channel_abnormal_list)
        # print('channel_list = ' + str(self.channel_list))

        self.set_HLayout_ton_Visable(self.channel_ton_list)

        self.mode_list = [
            self.ui.radio_mode0.isChecked(),
            self.ui.radio_mode1.isChecked(),
            self.ui.radio_mode2.isChecked(),
            self.ui.radio_mode3.isChecked(),
        ]
        self.current_mode_index = self.check_mode()  # 0: 只数据采集 1:异常检测 2：反馈控制
        # self.object_list = [
        #     self.ui.radio_metal.isChecked(),
        #     self.ui.radio_metalwire.isChecked(),
        #     self.ui.radio_metalp.isChecked(),
        #     self.ui.radio_paper.isChecked(),
        #     self.ui.radio_none.isChecked(),
        # ]
        self.ui.progressBar.setValue(20)  # 10%
        self._line.emit(20)
        self.channel_num = sum(self.channel_list)
        self._flag = 1
        daq_thead = threading.Thread(target=self._collect, kwargs={"channel_list": self.channel_list})
        daq_thead.setDaemon(True)
        daq_thead.start()
        print('connect successfully!')

    def set_HLayout_ton_Visable(self, channel_list):
        for HBox_ton in [self.label_ton_channel, self.spinBox_ton_real, self.spinBox_ton_calibration,
                         self.label_ton_calculated, self.spinBox_ton_range[:, 0], self.spinBox_ton_range[:, 1],
                         self.spinBox_ton_range[:, 2]]:
            for channel_index in range(len(channel_list)):
                HBox_ton[channel_index].setVisible(channel_list[channel_index])

    def runTime(self, channel, timesec, pass_num, samplate_num, sdk):
        rate = pass_num / samplate_num if samplate_num != 0 else 0
        channel = int(channel)
        self.model_display[channel][-1].setText("run time: " + str(timesec) + " sec")
        self.model_display[channel][-2].setText(str(round(rate * 100, 2)) + "%")
        self.model_display[channel][-2].setToolTip("pass: {0}\ntotal: {1}".format(pass_num, samplate_num))

    def check_mode(self):
        for i, enable in enumerate(self.mode_list):
            if enable:
                return i

    def change_mode(self, mode):
        if mode == 0:
            self.ui.radio_mode0.setChecked(True)
        if mode == 1:
            self.ui.radio_mode1.setChecked(True)
        if mode == 2:
            self.ui.radio_mode2.setChecked(True)
        if mode == 3:
            self.ui.radio_mode3.setChecked(True)

    def _loadModelSinger(self, i):
        # 更新模型
        model_name = "save_model_" + str(i) + ".pk"
        with open("model//" + model_name, "rb") as target:
            model = load(target)
            self.model[i] = model

    def updaterWin(self, curve=None, channel=None, y1=None, y2=None, y3=None, y4=None, res=None):
        # y1 y2 y3 y4  signal upper center down
        # print('updaterWin working')
        curve.clear()
        curve.enableAutoRange()
        curve.addLegend(offset=(-1, -1))
        QApplication.processEvents()
        upper = curve.plot(y2, pen=self.pen_white, name="upper/down")  # upper
        down = curve.plot(y4, pen=self.pen_white)  # down
        pen = self.pen_yellow if res == 0 else self.pen_red
        name = "good" if res == 0 else "NG"
        signal_ = curve.plot(y1, pen=pen, name=name)
        # # 超出区域 >
        curve.addItem(pg.FillBetweenItem(signal_, upper, (255, 0, 0, 100)))
        # # 超出区域 <
        curve.addItem(pg.FillBetweenItem(down, signal_, (255, 0, 0, 100)))
        curve.addItem(pg.FillBetweenItem(upper, down, (0, 250, 0, 150)))
        curve.plot(y3, pen=self.pen_black, name="center")  # center
        curve.plot(y1, pen=pen)

        if channel == "0":
            self.res += res
            self._loadModelSinger(0)
            self.show_result += 1

        elif channel == "1":
            self.res += res
            self._loadModelSinger(1)
            self.show_result += 1

        elif channel == "2":
            self.res += res
            self._loadModelSinger(2)
            self.show_result += 1

        elif channel == "3":
            self.res += res
            self._loadModelSinger(3)
            self.show_result += 1

        elif channel == "4":
            self.res += res
            self._loadModelSinger(4)
            self.show_result += 1

        elif channel == "5":
            self.res += res
            self._loadModelSinger(5)
            self.show_result += 1

        if self.show_result >= self.channel_num:
            self.show_result = 0
            if self.res > 0:
                self.res = 0
                self.result_list.append(-1)
                self._signal.emit("Error")
            else:
                self.res = 0
                self.result_list.append(1)
                self._signal.emit("Pass")
            self.score_plot2(self.result_list)
            # print('length of result list: ' + str(len(self.result_list)))

    def updater(self, curve, x, y, clean):
        # print('updater working')
        # x_axis_left, x_axis_right = self.ui.box_angle_1.value(), self.ui.box_angle_2.value()
        # _index = np.linspace(x_axis_left, x_axis_right, y.shape[0])
        # print(y.shape[0])
        if clean:
            curve.clear()
            curve.enableAutoRange()
        # curve.plot().setData(x, y)
        # curve.plot().setData(_index, y)
        curve.plot().setData(y)
        # QApplication.processEvents()  # 实
        # 时刷新页面

    def connect_sucessful(self):
        # 连接状态栏
        # for i in range(101):
        #     self.ui.progressBar.setValue(i)
        #     time.sleep(.001)
        self.ui.groupBox_5.setEnabled(False)
        self.ui.groupBox_4.setEnabled(False)
        # self.ui.groupBox_3.setEnabled(False)
        self.ui.btn_connect.setEnabled(False)
        self.ui.btn_reset.setEnabled(True)

        self.ui.checkBox_mode_change.setEnabled(True)
        self._signal.emit("daq is connect!")

    def stateChannel(self, state):
        return 0
        # if state == 0:
        #     self.ui.checkBox_ton_ch0.setCheckState(False)
        #     self.ui.checkBox_ton_ch1.setCheckState(False)
        #     self.ui.checkBox_ton_ch2.setCheckState(False)
        #     self.ui.checkBox_ton_ch3.setCheckState(False)
        #     self.ui.checkBox_ton_ch4.setCheckState(False)
        #     self.ui.checkBox_ton_ch5.setCheckState(False)
        #     self.ui.groupBox_model1.setEnabled(False)
        #     self.ui.groupBox_model2.setEnabled(False)
        #     self.ui.groupBox_model3.setEnabled(False)
        #     self.ui.groupBox_model4.setEnabled(False)
        #     self.ui.groupBox_model5.setEnabled(False)
        #     self.ui.groupBox_model6.setEnabled(False)
        # else:
        #     self.ui.checkBox_ton_ch0.setCheckState(True)
        #     self.ui.checkBox_ton_ch1.setCheckState(True)
        #     self.ui.checkBox_ton_ch2.setCheckState(True)
        #     self.ui.checkBox_ton_ch3.setCheckState(True)
        #     self.ui.checkBox_ton_ch4.setCheckState(True)
        #     self.ui.checkBox_ton_ch5.setCheckState(True)
        #     self.ui.groupBox_model1.setEnabled(True)
        #     self.ui.groupBox_model2.setEnabled(True)
        #     self.ui.groupBox_model3.setEnabled(True)
        #     self.ui.groupBox_model4.setEnabled(True)
        #     self.ui.groupBox_model5.setEnabled(True)
        #     self.ui.groupBox_model6.setEnabled(True)

    def _layoutSetting(self):
        # score
        self.scoreLayout = QVBoxLayout()
        self.score.showAxis("left", False)
        self.score.showAxis("bottom", False)
        self.score.showGrid(x=False, y=True)
        self.scoreLayout.addWidget(self.score)
        self.ui.graphicsView_3.setLayout(self.scoreLayout)
        # raw data
        [pw.showGrid(x=True, y=True) for pw in self.raw]
        self.gridLayout = QtWidgets.QGridLayout()
        self.gridLayout.addWidget(self.raw[0], 0, 1, 1, 1)
        self.gridLayout.addWidget(self.raw[1], 0, 2, 1, 1)
        self.gridLayout.addWidget(self.raw[2], 0, 3, 1, 1)
        self.gridLayout.addWidget(self.raw[3], 1, 1, 1, 1)
        self.gridLayout.addWidget(self.raw[4], 1, 2, 1, 1)
        self.gridLayout.addWidget(self.raw[5], 1, 3, 1, 1)
        self.ui.graphicsView.setLayout(self.gridLayout)
        # win data
        [pw.showGrid(x=True, y=True) for pw in self.win]
        self.gridLayout1 = QtWidgets.QGridLayout()
        self.gridLayout1.addWidget(self.win[0], 0, 1, 1, 1)
        self.gridLayout1.addWidget(self.win[1], 0, 2, 1, 1)
        self.gridLayout1.addWidget(self.win[2], 0, 3, 1, 1)
        self.gridLayout1.addWidget(self.win[3], 1, 1, 1, 1)
        self.gridLayout1.addWidget(self.win[4], 1, 2, 1, 1)
        self.gridLayout1.addWidget(self.win[5], 1, 3, 1, 1)
        self.ui.graphicsView_2.setLayout(self.gridLayout1)

    def _plot(self):
        # print('plotting')
        t1 = time.time()
        x_axis_left, x_axis_right = self.ui.box_angle_1.value(), self.ui.box_angle_2.value()
        self._index = np.linspace(x_axis_left, x_axis_right, self.data.shape[0])
        self.data.index = self._index
        [(pw.clear(), pw.enableAutoRange(), pw.plot(self.data.index, self.data.iloc[:, 0].values)) for pw in self.raw]
        t2 = time.time()
        # print(t2-t1)

    def score_plot1(self):
        # scatter
        # res = np.array(np.random.randint(0,2,20))
        res = np.array(np.random.random(20))
        x1 = np.argwhere(res > 0.7)
        self.score.plot([-2, 22], [0.7, 0.7], pen=self.pen_red, )
        self.score.plot(res, symbolBrush=(0, 255, 0))
        # self.score.plot(res, pen=self.pen_black)
        self.score.plot(x1.reshape(-1), res[x1].reshape(-1), symbolBrush=(255, 0, 0), )

    def score_plot2(self, l):
        # bar
        # res = np.array(np.random.random(30))
        res = np.array(l)[-30:]
        if len(res) < 30:
            res = np.concatenate((res, np.ones(30 - len(res))))
        x1 = np.argwhere(res > 0.5)
        x2 = np.argwhere(res <= 0.5)

        # print(res)
        # print(x1)
        # print(x2)

        bg1 = pg.BarGraphItem(x=x1, height=res[x1], width=1, brush='g', name="Good", )
        bg2 = pg.BarGraphItem(x=x2, height=res[x2], width=1, brush='r', name="Bad", )
        self.score.clear()
        self.score.enableAutoRange()
        # self.score.addLegend(offset=(1, -1))
        self.score.addItem(bg1)
        self.score.addItem(bg2)


if __name__ == "__main__":
    # 實例化一個應用
    app = QApplication(argv)
    # 實例化主窗口
    ui = parentWindow()
    ui.showMaximized()
    # ui.showNormal()
    # ui.show()
    exit(app.exec())
