import Api from '@/common/Api';
import { joinClientParams } from '@/common/request';
import config from '@/common/config';
import { randomUuid } from '@/common/utils/string';
import { getToken } from '@/common/utils/storage';
import { i18n } from '@/i18n';

const t = i18n.global.t;

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
    uploadFile: Api.newPost('/machines/{machineId}/files/{fileId}/upload?' + joinClientParams()),
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
    /** 上传ID（前端生成，保证唯一性） */
    uploadId: string;
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
    /** 相对路径（文件夹上传时使用） */
    relativePath?: string;
}

/**
 * 文件上传选项
 */
export interface UploadOptions {
    /** 进度回调 */
    onProgress?: (percent: number, uploadedSize: number, totalSize: number, speed: string) => void;
    /** 成功回调 */
    onSuccess?: () => void;
    /** 错误回调 */
    onError?: (error: Error) => void;
}

/**
 * 上传单个文件
 * @param file 文件对象
 * @param params 上传参数
 * @param options 上传选项
 * @returns Promise<void>
 */
export async function uploadFile(file: File, params: UploadParams, options: UploadOptions = {}): Promise<void> {
    const { onProgress, onSuccess, onError } = options;

    const formData = new FormData();
    formData.append('file', file);
    formData.append('uploadId', params.uploadId);
    formData.append('machineId', String(params.machineId));
    formData.append('authCertName', params.authCertName);
    formData.append('protocol', String(params.protocol));
    formData.append('fileId', String(params.fileId));
    formData.append('path', params.path);

    if (params.relativePath) {
        formData.append('relativePath', params.relativePath);
    }

    const token = getToken();
    const url = `${config.baseApiUrl}/machines/${params.machineId}/files/${params.fileId}/upload?token=${token}`;

    try {
        const xhr = new XMLHttpRequest();

        // 进度回调
        xhr.upload.onprogress = (event) => {
            if (event.lengthComputable && onProgress) {
                const percent = Math.round((event.loaded / event.total) * 100);
                const elapsed = (Date.now() - startTime) / 1000;
                const speedBytes = elapsed > 0 ? event.loaded / elapsed : 0;
                let speed = '0 B/s';
                if (speedBytes < 1024) {
                    speed = `${speedBytes.toFixed(0)} B/s`;
                } else if (speedBytes < 1024 * 1024) {
                    speed = `${(speedBytes / 1024).toFixed(1)} KB/s`;
                } else {
                    speed = `${(speedBytes / (1024 * 1024)).toFixed(1)} MB/s`;
                }
                onProgress(percent, event.loaded, event.total, speed);
            }
        };

        // 完成回调
        xhr.onload = () => {
            if (xhr.status === 200) {
                onSuccess?.();
            } else {
                onError?.(new Error(t('common.uploadFailed', { error: `HTTP ${xhr.status}` })));
            }
        };

        // 错误回调
        xhr.onerror = () => {
            onError?.(new Error(t('common.uploadFailed', { error: '网络错误' })));
        };

        const startTime = Date.now();
        xhr.open('POST', url);
        xhr.send(formData);
    } catch (error: any) {
        onError?.(new Error(t('common.uploadFailed', { error: error.message })));
    }
}

/**
 * 文件夹上传参数
 */
export interface FolderUploadParams {
    /** 上传ID（前端生成，保证唯一性） */
    uploadId: string;
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
 * 文件夹上传选项
 */
export interface FolderUploadOptions {
    /** 成功回调 */
    onSuccess?: () => void;
    /** 错误回调 */
    onError?: (error: Error) => void;
}

/**
 * 上传文件夹（使用 /upload-folder 接口）
 * @param files 文件列表
 * @param params 上传参数
 * @param options 上传选项
 * @returns Promise<void>
 */
export async function uploadFolder(files: FileList | File[], params: FolderUploadParams, options: FolderUploadOptions = {}): Promise<void> {
    const { onSuccess, onError } = options;

    const formData = new FormData();
    formData.append('uploadId', params.uploadId);
    formData.append('basePath', params.path);
    formData.append('machineId', String(params.machineId));
    formData.append('authCertName', params.authCertName);
    formData.append('protocol', String(params.protocol));

    // 添加所有文件
    const paths: string[] = [];
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const relativePath = (file as any).webkitRelativePath || file.name;
        formData.append('files', file);
        paths.push(relativePath);
    }

    // 添加路径数组
    paths.forEach((path) => {
        formData.append('paths', path);
    });

    const token = getToken();
    const url = `${config.baseApiUrl}/machines/${params.machineId}/files/${params.fileId}/upload-folder?token=${token}`;

    try {
        const xhr = new XMLHttpRequest();

        // 完成回调
        xhr.onload = () => {
            if (xhr.status === 200) {
                onSuccess?.();
            } else {
                onError?.(new Error(t('common.uploadFailed', { error: `HTTP ${xhr.status}` })));
            }
        };

        // 错误回调
        xhr.onerror = () => {
            onError?.(new Error(t('common.uploadFailed', { error: '网络错误' })));
        };

        xhr.open('POST', url);
        xhr.send(formData);
    } catch (error: any) {
        onError?.(new Error(t('common.uploadFailed', { error: error.message })));
    }
}
