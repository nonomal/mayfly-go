import Api from '@/common/Api';

export interface Session {
    sessionKey: string;
    title: string;
    createTime: string;
    updateTime: string;
}

export interface ToolCall {
    id: string;
    function: {
        name: string;
        arguments: string;
    };
}

export interface SessionMessage {
    turnId?: string;
    role: string;
    content: string;
    type?: string;
    time?: any;
    reasoningContent?: string;
    toolCalls?: ToolCall[];
    actionId?: string;
    extra?: any;
}

export const aiApi = {
    // 获取权限列表
    listSessions: Api.newGet<Session[]>('/ai/chat/sessions'),
    deleteSession: Api.newDelete('/ai/chat/sessions/{sessionKey}'),
    renameSession: Api.newPost('/ai/chat/sessions/rename'),
    listMessages: Api.newGet<SessionMessage[]>('/ai/chat/messages'),
};

export function getMachineTerminalSocketUrl(authCertName: any) {
    return `/machines/terminal/${authCertName}`;
}

export function getMachineRdpSocketUrl(authCertName: any) {
    return `/api/machines/rdp/${authCertName}`;
}
