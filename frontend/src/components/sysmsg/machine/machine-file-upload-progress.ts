import syssocket from '@/common/syssocket';
import { reactive, h } from 'vue';
import { ElNotification } from 'element-plus';
import MachineFileUploadProgress from './MachineFileUploadProgress.vue';

// 文件上传进度通知映射表（key: uploadId, value: 通知实例）
const fileUploadNotifyMap: Map<string, any> = new Map();

/**
 * 构建机器文件上传进度组件属性
 */
const buildMachineFileUploadProgressProps = (): any => {
    return {
        progress: reactive({
            filename: '',
            uploadedSize: 0,
            totalSize: 0,
            timestamp: 0,
            status: '', // '' | 'success' | 'exception'
        }),
    };
};

/**
 * 注册机器文件上传进度消息处理
 */
export async function registerMachineFileUploadProgress() {
    await syssocket.registerMsgHandler('machineFileUploadProgress', function (message: any) {
        const content = message.params;
        const uploadId = content.uploadId;

        // 上传完成或失败，关闭通知
        if (content.status === 'complete' || content.status === 'error') {
            const notify = fileUploadNotifyMap.get(uploadId);
            
            if (notify && notify.notification) {
                // 更新最终状态
                notify.props.progress.status = content.status === 'complete' ? 'success' : 'exception';
                notify.props.progress.percent = content.status === 'complete' ? 100 : notify.props.progress.percent;

                // 强制更新 VNode
                try {
                    if (notify.notification.state) {
                        notify.notification.state.message = h(MachineFileUploadProgress, notify.props);
                    } else if (notify.notification.vm) {
                        notify.notification.vm.exposed?.message?.(h(MachineFileUploadProgress, notify.props));
                    }
                } catch (e) {
                    console.warn('[MachineFileUpload] Failed to update notification VNode:', e);
                }

                // 1秒后关闭通知
                setTimeout(() => {
                    if (notify.notification) {
                        notify.notification.close();
                    }
                    fileUploadNotifyMap.delete(uploadId);
                }, 1000);
            }
            return;
        }

        // 获取或创建通知
        let notify = fileUploadNotifyMap.get(uploadId);
        if (!notify) {
            notify = {
                props: buildMachineFileUploadProgressProps(),
                notification: undefined,
            };
            fileUploadNotifyMap.set(uploadId, notify);
        }

        // 更新进度
        notify.props.progress.filename = content.filename || notify.props.progress.filename;
        notify.props.progress.uploadedSize = content.uploadedSize || 0;
        notify.props.progress.totalSize = content.totalSize || 0;
        notify.props.progress.timestamp = content.timestamp || 0;
        notify.props.progress.status = 'uploading';

        // 首次创建通知
        if (!notify.notification) {
            notify.notification = ElNotification({
                duration: 0,
                title: message.title || '机器文件上传',
                message: h(MachineFileUploadProgress, notify.props),
                showClose: true,
                offset: 60,
                customClass: 'machine-file-upload-notification',
            });
        } else {
            // 已存在通知，强制更新 message
            try {
                if (notify.notification.state) {
                    notify.notification.state.message = h(MachineFileUploadProgress, notify.props);
                } else if (notify.notification.vm) {
                    notify.notification.vm.exposed?.message?.(h(MachineFileUploadProgress, notify.props));
                }
            } catch (e) {
                console.warn('[MachineFileUpload] Failed to update notification VNode:', e);
            }
        }
    });
}
