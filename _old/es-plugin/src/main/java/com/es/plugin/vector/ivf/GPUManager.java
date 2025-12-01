package com.es.plugin.vector.ivf;

import jcuda.Pointer;
import jcuda.Sizeof;
import jcuda.runtime.JCuda;
import jcuda.runtime.cudaDeviceProp;

import static jcuda.runtime.JCuda.*;
import static jcuda.runtime.cudaError.cudaSuccess;

/**
 * GPU资源管理器
 * 负责GPU初始化、检测和资源管理
 */
public class GPUManager {

    private static GPUManager instance;
    private boolean gpuAvailable = false;
    private int deviceCount = 0;
    private int currentDevice = 0;
    private cudaDeviceProp deviceProperties;

    private GPUManager() {
        initializeGPU();
    }

    public static synchronized GPUManager getInstance() {
        if (instance == null) {
            instance = new GPUManager();
        }
        return instance;
    }

    /**
     * 初始化GPU
     */
    private void initializeGPU() {
        try {
            // 启用JCuda异常
            JCuda.setExceptionsEnabled(true);

            // 获取设备数量
            int[] count = new int[1];
            int result = cudaGetDeviceCount(count);

            if (result == cudaSuccess && count[0] > 0) {
                deviceCount = count[0];
                currentDevice = 0;

                // 获取设备属性
                deviceProperties = new cudaDeviceProp();
                cudaGetDeviceProperties(deviceProperties, currentDevice);

                gpuAvailable = true;

                System.out.println("=== GPU Initialization Successful ===");
                System.out.println("Device Count: " + deviceCount);
                System.out.println("Current Device: " + currentDevice);
                System.out.println("Device Name: " + deviceProperties.getName());
                System.out.println("Compute Capability: " +
                    deviceProperties.major + "." + deviceProperties.minor);
                System.out.println("Total Global Memory: " +
                    (deviceProperties.totalGlobalMem / (1024 * 1024)) + " MB");
                System.out.println("Multiprocessors: " + deviceProperties.multiProcessorCount);
                System.out.println("Max Threads Per Block: " + deviceProperties.maxThreadsPerBlock);
                System.out.println("=====================================");

            } else {
                gpuAvailable = false;
                System.out.println("No CUDA-capable GPU found. Falling back to CPU.");
            }

        } catch (Exception e) {
            gpuAvailable = false;
            System.err.println("GPU initialization failed: " + e.getMessage());
            System.out.println("Falling back to CPU computation.");
        }
    }

    /**
     * 检查GPU是否可用
     */
    public boolean isGPUAvailable() {
        return gpuAvailable;
    }

    /**
     * 获取GPU设备数量
     */
    public int getDeviceCount() {
        return deviceCount;
    }

    /**
     * 设置当前使用的GPU设备
     */
    public void setDevice(int deviceId) {
        if (deviceId >= 0 && deviceId < deviceCount) {
            cudaSetDevice(deviceId);
            currentDevice = deviceId;
            System.out.println("Switched to GPU device " + deviceId);
        } else {
            throw new IllegalArgumentException(
                "Invalid device ID: " + deviceId + ". Available devices: 0-" + (deviceCount - 1));
        }
    }

    /**
     * 获取当前GPU设备ID
     */
    public int getCurrentDevice() {
        return currentDevice;
    }

    /**
     * 获取设备属性
     */
    public cudaDeviceProp getDeviceProperties() {
        return deviceProperties;
    }

    /**
     * 获取GPU可用内存
     */
    public long getAvailableMemory() {
        if (!gpuAvailable) {
            return 0;
        }

        long[] free = new long[1];
        long[] total = new long[1];
        cudaMemGetInfo(free, total);

        return free[0];
    }

    /**
     * 同步GPU设备
     */
    public void synchronize() {
        if (gpuAvailable) {
            cudaDeviceSynchronize();
        }
    }

    /**
     * 重置GPU设备
     */
    public void resetDevice() {
        if (gpuAvailable) {
            cudaDeviceReset();
            System.out.println("GPU device reset completed.");
        }
    }
}
