import Api from '@/common/Api';
import { registerSqlExecAborter, createSqlExecNotification } from '@/components/sysmsg/db/db-sql-exec-progress';
import { AesEncrypt } from '@/common/crypto';
import { joinClientParams } from '@/common/request';

export const dbApi = {
    // 获取权限列表
    dbs: Api.newGet('/dbs'),
    dbTags: Api.newGet('/dbs/tags'),
    saveDb: Api.newPost('/dbs'),
    deleteDb: Api.newDelete('/dbs/{id}'),
    dumpDb: Api.newPost('/dbs/{id}/dump'),
    tableInfos: Api.newGet('/dbs/{id}/t-infos'),
    tableIndex: Api.newGet('/dbs/{id}/t-index'),
    tableDdl: Api.newGet('/dbs/{id}/t-create-ddl'),
    copyTable: Api.newPost('/dbs/{id}/copy-table'),
    columnMetadata: Api.newGet('/dbs/{id}/c-metadata'),
    pgSchemas: Api.newGet('/dbs/{id}/pg/schemas'),
    // 获取表即列提示
    hintTables: Api.newGet('/dbs/{id}/hint-tables'),
    sqlExec: Api.newPost('/dbs/{id}/exec-sql').withBeforeHandler(async (param: any) => await encryptField(param, 'sql')),
    // 保存sql
    saveSql: Api.newPost('/dbs/{id}/sql'),
    // 获取保存的sql
    getSql: Api.newGet('/dbs/{id}/sql'),
    // 获取保存的sql names
    getSqlNames: Api.newGet('/dbs/{id}/sql-names'),
    deleteDbSql: Api.newDelete('/dbs/{id}/sql'),
    // 获取数据库sql执行记录
    getSqlExecs: Api.newGet('/dbs/sql-execs'),
    // 获取数据库兼容版本
    getCompatibleDbVersion: Api.newGet('/dbs/{id}/version'),

    instances: Api.newGet('/instances'),
    getInstance: Api.newGet('/instances/{instanceId}'),
    getAllDatabase: Api.newPost('/instances/databases'),
    getDbNamesByAc: Api.newGet('/instances/databases/{authCertName}'),
    getInstanceServerInfo: Api.newGet('/instances/{instanceId}/server-info'),
    testConn: Api.newPost('/instances/test-conn'),
    saveInstance: Api.newPost('/instances'),
    deleteInstance: Api.newDelete('/instances/{id}'),

    // 获取数据库备份列表
    getDbBackups: Api.newGet('/dbs/{dbId}/backups'),
    createDbBackup: Api.newPost('/dbs/{dbId}/backups'),
    deleteDbBackup: Api.newDelete('/dbs/{dbId}/backups/{backupId}'),
    getDbNamesWithoutBackup: Api.newGet('/dbs/{dbId}/db-names-without-backup'),
    enableDbBackup: Api.newPut('/dbs/{dbId}/backups/{backupId}/enable'),
    disableDbBackup: Api.newPut('/dbs/{dbId}/backups/{backupId}/disable'),
    startDbBackup: Api.newPut('/dbs/{dbId}/backups/{backupId}/start'),
    saveDbBackup: Api.newPut('/dbs/{dbId}/backups/{id}'),
    getDbBackupHistories: Api.newGet('/dbs/{dbId}/backup-histories'),
    restoreDbBackupHistory: Api.newPost('/dbs/{dbId}/backup-histories/{backupHistoryId}/restore'),
    deleteDbBackupHistory: Api.newDelete('/dbs/{dbId}/backup-histories/{backupHistoryId}'),

    // 获取数据库恢复列表
    getDbRestores: Api.newGet('/dbs/{dbId}/restores'),
    createDbRestore: Api.newPost('/dbs/{dbId}/restores'),
    deleteDbRestore: Api.newDelete('/dbs/{dbId}/restores/{restoreId}'),
    getDbNamesWithoutRestore: Api.newGet('/dbs/{dbId}/db-names-without-restore'),
    enableDbRestore: Api.newPut('/dbs/{dbId}/restores/{restoreId}/enable'),
    disableDbRestore: Api.newPut('/dbs/{dbId}/restores/{restoreId}/disable'),
    saveDbRestore: Api.newPut('/dbs/{dbId}/restores/{id}'),
};

export const dbSqlExecApi = {
    // 根据业务key获取sql执行信息
    getSqlExecByBizKey: Api.newGet('/dbs/sql-execs'),
};
export const encryptField = async (param: any, field: string) => {
    // sql编码处理
    if (!param['_encrypted'] && param[field]) {
        // 判断是开发环境就打印sql
        if (process.env.NODE_ENV === 'development') {
            console.log(param[field]);
        }
        // 使用aes加密sql
        param['_encrypted'] = 1;
        param[field] = AesEncrypt(param[field]);
        // console.log('解密结果', DesDecrypt(param[field]));
    }
    return param;
};

/**
 * 上传SQL文件并执行
 * @param file 文件对象
 * @param params 上传参数
 * @param options 上传选项
 * @returns { uploadId: string; abort: () => void } 返回包含 uploadId 和中止方法的对象
 */
export function uploadSqlFile(
    file: File,
    params: {
        dbId: number;
        dbName: string;
    },
    options: {
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    } = {}
): { uploadId: string; abort: () => void } {
    // 生成 uploadId
    const uploadId = `sql_exec_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;

    // 使用 URLSearchParams 构建查询参数
    const queryParams = new URLSearchParams({
        db: params.dbName,
        uploadId: uploadId,
        filename: file.name,
    }).toString();

    // 创建 Api 实例
    const api = Api.newPost(`/dbs/${params.dbId}/exec-sql-file`);

    // 使用 uploadRaw 直接传递文件流
    const { abort } = api.uploadRaw(file, queryParams, {
        onSuccess: () => {
            options.onSuccess?.();
        },
        onError: (error) => {
            options.onError?.(error);
        },
    });

    // 创建SQL执行进度通知
    createSqlExecNotification(uploadId, {
        id: uploadId,
        title: file.name,
        dbCode: '',
        dbName: params.dbName,
        executedStatements: 0,
        terminated: false,
        status: 'uploading',
        clientId: '',
    });

    // 注册取消器（在获取到abort方法后）
    registerSqlExecAborter(uploadId, abort);

    return { uploadId, abort };
}
