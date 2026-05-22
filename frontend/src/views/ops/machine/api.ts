import Api, { UploadOptions } from '@/common/Api';
import { createUploadFileNotification, registerUploadFileAborter } from '@/components/sysmsg/machine/machine-file-upload-progress';
import { createUploadFolderNotification, registerUploadFolderAborter } from '@/components/sysmsg/machine/machine-folder-upload-progress';

export const machineApi = {
    // 获取权限列表
    list: Api.newGet('/machines'),
    getByCodes: Api.newGet('/machines/simple'),
    tagList: Api.newGet('/machines/tags'),
    getMachinePwd: Api.newGet('/machines/{id}/pwd'),
    info: Api.newGet('/machines/{id}/sysinfo'),
    stats: Api.newGet('/machines/{id}/stats'),
    process: Api.newGet('/machines/{id}/process'),
    // 终止进程
    killProcess: Api.newDelete('/machines/{id}/process'),
    users: Api.newGet('/machines/{id}/users'),
    groups: Api.newGet('/machines/{id}/groups'),
    testConn: Api.newPost('/machines/test-conn'),
    // 保存按钮
    saveMachine: Api.newPost('/machines'),
    // 调整状态
    changeStatus: Api.newPut('/machines/{id}/{status}'),
    // 删除机器
    del: Api.newDelete('/machines/{id}'),
    scripts: Api.newGet('/machines/{machineId}/scripts'),
    scriptCategorys: Api.newGet('/machines/scripts/categorys'),
    runScript: Api.newGet('/machines/scripts/{scriptId}/{ac}/run'),
    saveScript: Api.newPost('/machines/{machineId}/scripts'),
    deleteScript: Api.newDelete('/machines/{machineId}/scripts/{scriptId}'),
    // 获取配置文件列表
    files: Api.newGet('/machines/{id}/files'),
    lsFile: Api.newGet('/machines/{machineId}/files/{fileId}/read-dir'),
    dirSize: Api.newGet('/machines/{machineId}/files/{fileId}/dir-size'),
    fileStat: Api.newGet('/machines/{machineId}/files/{fileId}/file-stat'),
    rmFile: Api.newPost('/machines/{machineId}/files/{fileId}/remove'),
    cpFile: Api.newPost('/machines/{machineId}/files/{fileId}/cp'),
    renameFile: Api.newPost('/machines/{machineId}/files/{fileId}/rename'),
    mvFile: Api.newPost('/machines/{machineId}/files/{fileId}/mv'),
    uploadFile: Api.newUpload('/machines/{machineId}/files/{fileId}/upload'),
    uploadFolder: Api.newPost('/machines/{machineId}/files/{fileId}/upload-folder'),
    fileContent: Api.newGet('/machines/{machineId}/files/{fileId}/read'),
    downloadFile: Api.newGet('/machines/{machineId}/files/{fileId}/download'),
    createFile: Api.newPost('/machines/{machineId}/files/{id}/create-file'),
    // 修改文件内容
    updateFileContent: Api.newPost('/machines/{machineId}/files/{id}/write'),
    // 添加文件or目录
    addConf: Api.newPost('/machines/{machineId}/files'),
    // 删除配置的文件or目录
    delConf: Api.newDelete('/machines/{machineId}/files/{id}'),
    // 机器终端操作记录列表
    termOpRecs: Api.newGet('/machines/{machineId}/term-recs'),
};

export const cronJobApi = {
    list: Api.newGet('/machine-cronjobs'),
    relateMachineIds: Api.newGet('/machine-cronjobs/machine-ids'),
    relateCronJobIds: Api.newGet('/machine-cronjobs/cronjob-ids'),
    save: Api.newPost('/machine-cronjobs'),
    delete: Api.newDelete('/machine-cronjobs/{id}'),
    run: Api.newPost('/machine-cronjobs/run/{key}'),
    execList: Api.newGet('/machine-cronjobs/execs'),
};

export const cmdConfApi = {
    list: Api.newGet('/machine/security/cmd-confs'),
    save: Api.newPost('/machine/security/cmd-confs'),
    delete: Api.newDelete('/machine/security/cmd-confs/{id}'),
};

/**
 * 获取终端 WebSocket URL
 */
export function getMachineTerminalSocketUrl(authCertName: any) {
    return `/machines/terminal/${authCertName}`;
}

/**
 * 获取 RDP WebSocket URL
 */
export function getMachineRdpSocketUrl(authCertName: any) {
    return `/api/machines/rdp/${authCertName}`;
}

/**
 * 文件上传参数
 */
export interface UploadParams {
    /** 上传ID（可选，不传则内部自动生成） */
    uploadId?: string;
    /** 机器ID */
    machineId: number;
    /** 认证证书名称 */
    authCertName: string;
    /** 协议类型 */
    protocol: number;
    /** 文件ID */
    fileId: number;
    /** 目标路径 */
    path: string;
    /** 文件名 */
    filename: string;

    /** 是否创建进度通知 */
    createProgressNotify?: boolean;
    /** 是否是文件夹上传的一部分 */
    isFolderUpload?: boolean;
}

/**
 * 上传单个文件
 * @param file 文件对象
 * @param params 上传参数
 * @param options 上传选项
 * @returns { uploadId: string; abort: () => void } 返回包含 uploadId 和中止方法的对象
 */
export function uploadFile(file: File, params: UploadParams, options: UploadOptions = {}): { uploadId: string; abort: () => void } {
    // 业务层生成 uploadId
    const uploadId = params.uploadId || `upload_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;

    // 构建查询参数
    const queryParams = new URLSearchParams({
        machineId: String(params.machineId),
        authCertName: params.authCertName,
        protocol: String(params.protocol),
        fileId: String(params.fileId),
        path: params.path,
        uploadId: uploadId,
        filename: file.name,
    });

    // 如果是文件夹上传，添加标识参数
    if (params.isFolderUpload) {
        queryParams.set('isFolderUpload', 'true');
    }

    // 直接使用文件流作为 body，不包装为 FormData
    const { abort } = machineApi.uploadFile.uploadRaw(file, queryParams.toString(), {
        ...options,
    });

    if (params.createProgressNotify !== false) {
        createUploadFileNotification(uploadId, {
            authCertName: params.authCertName,
            path: params.path,
            filename: file.name,
        });

        // 注册取消方法
        registerUploadFileAborter(uploadId, abort);
    }

    return { uploadId, abort };
}

/**
 * 文件夹上传参数
 */
export interface FolderUploadParams {
    /** 上传ID（可选，不传则内部自动生成） */
    uploadId?: string;
    /** 机器ID */
    machineId: number;
    /** 认证证书名称 */
    authCertName: string;
    /** 协议类型 */
    protocol: number;
    /** 文件ID */
    fileId: number;
    /** 目标路径 */
    path: string;
}

/**
 * 上传文件夹（逐个文件流式上传，保持目录结构）
 */
export function uploadFolder(files: FileList | File[], params: FolderUploadParams, options: UploadOptions = {}): { uploadId: string; abort: () => void } {
    const uploadId = params.uploadId || `folder_${params.fileId}_${Date.now()}`;
    const fileArray = Array.from(files);
    const totalFiles = fileArray.length;
    const totalSize = fileArray.reduce((sum, file) => sum + file.size, 0);
    let isAborted = false;
    const abortControllers: (() => void)[] = []; // 存储所有正在进行的上传的取消方法

    const filepaths: string[] = [];
    // 创建上传任务
    const uploadTasks = fileArray.map((file, index) => {
        const relativePath = (file as any).webkitRelativePath || file.name;
        const fullPath = `${params.path}/${relativePath}`;
        const dirPath = fullPath.substring(0, fullPath.lastIndexOf('/'));

        filepaths.push(fullPath);

        return () =>
            new Promise<void>((resolve, reject) => {
                console.log(`[FolderUpload] 开始上传 ${index + 1}/${totalFiles}:`, fullPath);

                const { abort } = uploadFile(
                    file,
                    {
                        uploadId,
                        machineId: params.machineId,
                        authCertName: params.authCertName,
                        protocol: params.protocol,
                        fileId: params.fileId,
                        path: dirPath,
                        filename: file.name,
                        isFolderUpload: true,
                        createProgressNotify: false,
                    },
                    {
                        onSuccess: () => {
                            console.log(`[FolderUpload] 上传成功 ${index + 1}/${totalFiles}:`, fullPath);
                            resolve();
                        },
                        onError: (error) => {
                            console.log(`[FolderUpload] 上传失败 ${index + 1}/${totalFiles}:`, fullPath, error.message);
                            if (error.name === 'AbortError' || isAborted) {
                                reject(error);
                            } else {
                                options.onError?.(error);
                                resolve();
                            }
                        },
                    }
                );

                // 将当前文件的取消方法添加到列表中
                abortControllers.push(abort);
            });
    });

    // 初始化进度通知
    const folderName = fileArray[0]?.webkitRelativePath?.split('/')[0] || 'folder';
    createUploadFolderNotification(uploadId, {
        authCertName: params.authCertName,
        path: params.path,
        folderName,
        totalFiles,
        totalSize,
        uploadedSize: 0,
        uploadedFiles: 0,
        finishedFiles: 0,
        status: 'uploading',
        files: new Map(filepaths.map((filepath) => [filepath, { path: filepath, status: 'waiting', progress: 0, currentSize: 0, totalSize: 0, timestamp: 0 }])),
    });

    // 并发执行
    const executeUploads = async () => {
        const maxConcurrent = 3;
        const running = new Set<Promise<void>>();
        let index = 0;

        const launchNext = () => {
            if (index >= uploadTasks.length || isAborted) return;

            const task = uploadTasks[index++]();
            running.add(task);

            task.finally(() => {
                running.delete(task);
                if (!isAborted) launchNext();
            });
        };

        // 启动初始并发
        for (let i = 0; i < maxConcurrent && i < uploadTasks.length; i++) {
            launchNext();
        }

        // 等待所有完成
        while (running.size > 0) {
            await Promise.race(running);
        }

        if (!isAborted) {
            console.log('[FolderUpload] 全部完成');
            options.onSuccess?.();
        }
    };

    executeUploads();

    const res = {
        uploadId,
        abort: () => {
            isAborted = true;
            // 取消所有正在进行的上传请求
            abortControllers.forEach((abort) => abort());
        },
    };

    registerUploadFolderAborter(uploadId, res.abort);
    return res;
}
