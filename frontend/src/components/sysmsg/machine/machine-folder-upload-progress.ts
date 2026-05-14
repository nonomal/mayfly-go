import { ElNotification } from 'element-plus';
import { h, reactive } from 'vue';
import MachineFolderUploadProgress from '@/components/sysmsg/machine/MachineFolderUploadProgress.vue';
import syssocket from '@/common/syssocket';

// 文件夹上传通知 Map
const folderUploadNotifyMap = new Map<string, any>();

/**
 * 构建文件夹上传进度组件的 props
 */
const buildMachineFolderUploadProgressProps = (): any => {
    return {
        progress: reactive({
            folderName: '',
            totalFiles: 0,
            uploadedFiles: 0,
            totalSize: 0,
            uploadedSize: 0,
            lastFile: '',
            uploadingFiles: [] as string[],
            timestamp: 0,
            status: '',
        }),
    };
};

/**
 * 注册文件夹上传进度消息处理
 */
export async function registerFolderUploadProgressHandler() {
    await syssocket.registerMsgHandler('machineFolderUploadProgress', function (message: any) {
        const content = message.params;
        const uploadId = content.uploadId;

        if (!uploadId) {
            return;
        }

        // 获取或创建通知
        let notify = folderUploadNotifyMap.get(uploadId);

        if (!notify) {
            // 首次创建通知
            const props = buildMachineFolderUploadProgressProps();
            
            const notificationInstance = ElNotification({
                title: '文件夹上传',
                message: h(MachineFolderUploadProgress, props),
                duration: 0,
                position: 'top-right',
                offset: 60,
                customClass: 'machine-folder-upload-notify',
                onClose: () => {
                    folderUploadNotifyMap.delete(uploadId);
                },
            });

            notify = {
                props,
                notification: notificationInstance,
            };

            folderUploadNotifyMap.set(uploadId, notify);
        }

        // 上传完成或失败，关闭通知
        if (content.status === 'complete' || content.status === 'error') {
            if (notify && notify.notification) {
                // 更新最终状态
                notify.props.progress.status = content.status === 'complete' ? 'success' : 'exception';

                // 强制更新 VNode
                try {
                    if (notify.notification.state) {
                        notify.notification.state.message = h(MachineFolderUploadProgress, notify.props);
                    }
                } catch (e) {
                    console.warn('[MachineFolderUpload] Failed to update notification VNode:', e);
                }

                // 1秒后关闭通知
                setTimeout(() => {
                    if (notify.notification) {
                        notify.notification.close();
                    }
                    folderUploadNotifyMap.delete(uploadId);
                }, 1000);
            }
            return;
        }

        // 更新进度
        if (content.status === 'uploading') {
            notify.props.progress.folderName = content.folderName || '';
            notify.props.progress.totalFiles = content.totalFiles || 0;
            notify.props.progress.uploadedFiles = content.uploadedFiles || 0;
            notify.props.progress.totalSize = content.totalSize || 0;
            notify.props.progress.uploadedSize = content.uploadedSize || 0;
            notify.props.progress.lastFile = content.lastFile || '';
            notify.props.progress.uploadingFiles = content.uploadingFiles || [];
            notify.props.progress.timestamp = content.timestamp || 0;
            notify.props.progress.status = 'uploading';

            // 强制更新 VNode
            try {
                if (notify.notification && notify.notification.state) {
                    notify.notification.state.message = h(MachineFolderUploadProgress, notify.props);
                }
            } catch (e) {
                console.warn('[MachineFolderUpload] Failed to update notification VNode:', e);
            }
        }
    });
}
