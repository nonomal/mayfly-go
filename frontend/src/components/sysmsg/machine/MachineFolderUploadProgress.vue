<template>
    <div class="machine-folder-upload-progress">
        <!-- 文件夹信息 -->
        <div class="progress-header">
            <span class="folder-name">{{ progress.folderName }}</span>
            <span class="file-count">{{ progress.uploadedFiles }}/{{ progress.totalFiles }}</span>
        </div>
        
        <!-- 整体进度条 -->
        <el-progress
            :percentage="percent"
            :status="progress.status"
            :stroke-width="10"
        />
        
        <!-- 整体进度信息 -->
        <div class="progress-info">
            <span class="size-info">{{ formatSize(progress.uploadedSize) }} / {{ formatSize(progress.totalSize) }}</span>
            <span class="percent">{{ percent }}%</span>
        </div>

        <!-- 正在上传的文件列表 -->
        <div v-if="progress.uploadingFiles && progress.uploadingFiles.length > 0" class="uploading-files">
            <div class="section-title">正在上传 ({{ progress.uploadingFiles.length }} 个并发):</div>
            <div v-for="(file, index) in progress.uploadingFiles" :key="index" class="uploading-file">
                <el-icon class="loading-icon"><Loading /></el-icon>
                <span class="file-path">{{ file }}</span>
            </div>
        </div>

        <!-- 最后完成的文件 -->
        <div v-if="progress.lastFile && progress.status === 'uploading'" class="last-file">
            <el-icon class="success-icon"><Check /></el-icon>
            <span class="file-path">{{ progress.lastFile }}</span>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { Loading, Check } from '@element-plus/icons-vue';
import { computed } from 'vue';

const props = defineProps({
    progress: {
        type: Object,
        required: true,
    },
});

// 计算百分比
const percent = computed(() => {
    if (!props.progress.totalSize || !props.progress.uploadedSize) {
        return 0;
    }
    return Math.min(100, Math.floor((props.progress.uploadedSize / props.progress.totalSize) * 100));
});

// 格式化文件大小
const formatSize = (bytes: number): string => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i];
};
</script>

<style lang="scss" scoped>
.machine-folder-upload-progress {
    padding: 8px 0;
    max-width: 500px;
    
    .progress-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;
        
        .folder-name {
            font-weight: 600;
            font-size: 14px;
            color: var(--el-text-color-primary);
        }
        
        .file-count {
            font-size: 12px;
            color: var(--el-text-color-secondary);
        }
    }
    
    .progress-info {
        margin-top: 6px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        
        .size-info {
            font-size: 12px;
            color: var(--el-text-color-secondary);
        }
        
        .percent {
            font-size: 12px;
            font-weight: 600;
            color: var(--el-text-color-primary);
        }
    }
    
    .uploading-files {
        margin-top: 12px;
        padding-top: 12px;
        border-top: 1px solid var(--el-border-color-lighter);
        
        .section-title {
            font-size: 12px;
            font-weight: 600;
            color: var(--el-color-primary);
            margin-bottom: 8px;
        }
        
        .uploading-file {
            display: flex;
            align-items: center;
            gap: 6px;
            padding: 4px 0;
            font-size: 11px;
            color: var(--el-text-color-regular);
            
            .loading-icon {
                animation: rotating 2s linear infinite;
                color: var(--el-color-primary);
            }
            
            .file-path {
                flex: 1;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        }
    }
    
    .last-file {
        margin-top: 8px;
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 6px 8px;
        background: var(--el-color-success-light-9);
        border-radius: 4px;
        font-size: 11px;
        
        .success-icon {
            color: var(--el-color-success);
        }
        
        .file-path {
            color: var(--el-color-success);
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
    }
}

@keyframes rotating {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}
</style>
