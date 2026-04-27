export interface IMilvus {
    id: number;
    code: string;
    name: string;
    host: string;
    username?: string;
    password?: string;
    database?: string;
    sshTunnelMachineId?: number;
    createTime?: string;
}

export interface IDatabase {
    name: string;
}

export interface ICollection {
    name: string;
    description?: string;
    loaded?: boolean;
    shardsNum?: number;
}

export interface IPartition {
    collectionName: string;
    name: string;
}

export interface IIndex {
    collectionName: string;
    fieldName: string;
    indexName: string;
    indexType: string;
    metricType: string;
}

export interface IUser {
    username: string;
}

export interface IRole {
    name: string;
}

export interface IResourceGroup {
    name: string;
}

export interface IPrivilegeGroup {
    GroupName: string;
    Privileges: string[];
}
