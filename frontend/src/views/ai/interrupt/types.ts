/**
 * 中断信息组件统一类型定义
 */

import type { SessionMessage } from '../api';

/**
 * 内部消息类型（包含 extra 字段）
 */
export type InternalMessage = SessionMessage & {
    codeId?: string;
    extra?: {
        type?: 'interrupt' | 'notification' | string;
        content?: any;
        toolStatus?: string;
        interruptId?: string;
        [key: string]: any;
    };
};

/**
 * 中断操作事件数据
 */
export interface InterruptActionEvent {
    turnId: string; // 轮次Id
    interruptId: string; // 中断ID
    interruptType: string; // 中断类型（如 'APPROVAL', 'PARAM_COMPLETION' 等）
    action: string; // 操作（如 'approve', 'reject', 'complete', 'cancel' 等）
    payload?: any; // 操作携带的额外数据
}

/**
 * 中断组件事件处理器类型
 */
export type InterruptActionHandler = (action: InterruptActionEvent) => void | Promise<void>;

/**
 * 中断组件 Props 接口
 * 所有中断组件必须遵循此接口
 */
export interface InterruptComponentProps {
    data: InternalMessage; // 完整的内部消息对象
    readonly?: boolean; // 是否只读模式
}
