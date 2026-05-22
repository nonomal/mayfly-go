import { getToken } from '@/common/utils/storage';

import { createWebSocket } from './request';
import { ElNotification } from 'element-plus';
import { MsgSubtypeEnum } from './commonEnum';
import EnumValue from './Enum';
import { h } from 'vue';
import { MessageRenderer } from '@/components/message/message';
import { initMachineSysMsgs } from '@/components/sysmsg/machine';
import { initDbSysMsgs } from '@/components/sysmsg/db';

/**
 * 初始化全局系统消息
 */
export function initSysMsgs() {
    initMachineSysMsgs();
    initDbSysMsgs();
}

class SysSocket {
    /**
     * socket连接
     */
    socket: WebSocket | null = null;

    /**
     * key -> 消息类别，value -> 消息对应的处理器函数
     */
    categoryHandlers: Map<string, any> = new Map();

    /**
     * 重连定时器
     */
    reconnectTimer: number | null = null;

    /**
     * 当前重连次数
     */
    reconnectCount: number = 0;

    /**
     * 基础重连延迟（毫秒）
     */
    baseReconnectDelay: number = 3000;

    /**
     * 是否正在重连
     */
    isReconnecting: boolean = false;

    /**
     * 是否手动关闭
     */
    isManualClose: boolean = false;

    /**
     * 初始化全局系统消息websocket
     */
    async init() {
        // 存在则不需要重新建立连接
        if (this.socket) {
            return;
        }
        const token = getToken();
        if (!token) {
            return null;
        }
        console.log('init system ws');
        try {
            this.isManualClose = false;
            await this.connect();
        } catch (e) {
            console.error('open system ws error', e);
        }
    }

    /**
     * 建立 WebSocket 连接
     */
    private async connect(): Promise<void> {
        this.socket = await createWebSocket('/sysmsg');
        
        this.socket.onopen = () => {
            console.log('WebSocket connected');
            this.resetReconnect();
        };

        this.socket.onmessage = async (event: { data: string }) => {
            let message;
            try {
                message = JSON.parse(event.data);
            } catch (e) {
                console.error('解析ws消息失败', e);
                return;
            }

            // 存在消息类别对应的处理器，则进行处理，否则进行默认通知处理
            const handler = this.categoryHandlers.get(message.category);
            if (handler) {
                handler(message);
                return;
            }

            const msgSubtype = EnumValue.getEnumByValue(MsgSubtypeEnum, message.subtype);
            if (!msgSubtype) {
                console.log(`not found msg subtype: ${message.subtype}`);
                return;
            }

            // 动态导入 i18n 或延迟获取 i18n 实例
            let title = '';
            try {
                // 方式1: 动态导入
                const { i18n } = await import('@/i18n');
                title = i18n.global.t(msgSubtype?.label);
            } catch (e) {
                console.warn('i18n not ready, using default title');
            }

            ElNotification({
                duration: 0,
                title,
                message: h(MessageRenderer, { content: message.msg }),
                type: msgSubtype?.extra.notifyType || 'info',
            });
        };

        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        this.socket.onclose = (event) => {
            console.log('WebSocket closed:', event.code, event.reason);
            this.socket = null;
            
            // 如果不是手动关闭，则尝试重连
            if (!this.isManualClose) {
                this.handleReconnect();
            }
        };
    }

    /**
     * 处理重连逻辑
     */
    private handleReconnect() {
        if (this.isReconnecting) {
            return;
        }

        this.isReconnecting = true;
        this.reconnectCount++;

        // 固定延迟重连策略：每 3 秒重试一次
        const delay = this.baseReconnectDelay;
        console.log(`WebSocket 将在 ${delay}ms 后尝试第 ${this.reconnectCount} 次重连`);

        this.reconnectTimer = window.setTimeout(async () => {
            this.isReconnecting = false;
            try {
                const token = getToken();
                if (!token) {
                    console.warn('Token 不存在，停止重连');
                    return;
                }
                console.log(`尝试第 ${this.reconnectCount} 次重连 WebSocket`);
                await this.connect();
            } catch (e) {
                console.error(`第 ${this.reconnectCount} 次重连失败:`, e);
                this.handleReconnect();
            }
        }, delay);
    }

    /**
     * 重置重连状态
     */
    private resetReconnect() {
        this.isReconnecting = false;
        this.reconnectCount = 0;
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }
    }

    destory() {
        this.isManualClose = true;
        this.resetReconnect();
        this.socket?.close();
        this.socket = null;
        this.categoryHandlers?.clear();
    }

    /**
     * 注册消息处理函数
     *
     * @param category 消息类别
     * @param handlerFunc 消息处理函数
     */
    async registerMsgHandler(category: any, handlerFunc: any) {
        if (this.categoryHandlers.has(category)) {
            console.log(`${category}该类别消息处理器已存在...`);
            return;
        }
        if (typeof handlerFunc != 'function') {
            throw new Error('message handler需为函数');
        }
        this.categoryHandlers.set(category, handlerFunc);
    }
}

// 全局系统消息websocket;
const sysSocket = new SysSocket();

export default sysSocket;
