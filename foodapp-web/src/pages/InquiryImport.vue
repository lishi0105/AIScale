<template>
  <div class="inquiry-import">
    <div class="upload-section">
        <div class="tips">
          <el-alert
            title="导入说明"
            type="info"
            :closable="false"
            show-icon
            class="alert-compact"
          >
            <p>1. 请上传标准格式的Excel文件（.xlsx 或 .xls）</p>
            <p>2. 如果存在相同标题或日期的询价单，系统会提示是否覆盖</p>
            <p>3. 导入过程会在后台执行，您可以查看实时进度</p>
          </el-alert>
        </div>

        <el-form :model="form" label-width="120px" style="margin-top: 20px;">
          <el-form-item label="Excel文件" required>
            <el-upload
              ref="uploadRef"
              :auto-upload="false"
              :limit="1"
              :on-change="handleFileChange"
              :on-remove="handleFileRemove"
              :file-list="fileList"
              accept=".xlsx,.xls"
              drag
            >
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">
                将文件拖到此处，或<em>点击上传</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  仅支持 .xlsx 或 .xls 格式的Excel文件
                </div>
              </template>
            </el-upload>
          </el-form-item>

          <el-form-item class="button-group">
            <el-button 
              type="primary" 
              @click="handleSubmit"
              :loading="uploading"
              :disabled="!canSubmit"
            >
              {{ uploading ? '上传中...' : '开始导入' }}
            </el-button>
            <el-button @click="handleReset" :disabled="uploading || !!importTask">重置</el-button>
            <el-button @click="handleCancel" :disabled="uploading || !!importTask">取消</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 进度显示 -->
      <div v-if="importTask" class="progress-section">
        <el-divider />
        <h3>导入进度</h3>
        
        <div class="task-info">
          <div class="task-details">
            <p><strong>文件名：</strong>{{ importTask.file_name }}</p>
            <p><strong>状态：</strong>
              <el-tag :type="statusType">{{ statusText }}</el-tag>
            </p>
          </div>

          <div style="margin-top: 16px;">
            <el-progress 
              :percentage="importTask.progress" 
              :status="progressStatus"
              :stroke-width="20"
            />
          </div>

          <!-- 成功信息 -->
          <el-alert
            v-if="importTask.status === 'success'"
            title="导入成功！"
            type="success"
            :closable="true"
            show-icon
            style="margin-top: 16px;"
            @close="handleCloseAlert"
          >
            <p v-if="importTask.inquiry_id">询价单ID: {{ importTask.inquiry_id }}</p>
          </el-alert>

          <!-- 失败信息 -->
          <el-alert
            v-if="importTask.status === 'failed'"
            title="导入失败"
            type="error"
            :closable="true"
            show-icon
            style="margin-top: 16px;"
            @close="handleCloseAlert"
          >
            <p>{{ importTask.error_message || '未知错误' }}</p>
          </el-alert>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile, UploadInstance } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { InquiryAPI, type ImportError, type ImportTask } from '@/api/inquiry'

// Props
interface Props {
  orgId: string
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  close: []
  importSuccess: []
}>()

// 监听弹窗打开，重置状态
import { watch } from 'vue'
watch(() => props.orgId, () => {
  if (props.orgId) {
    resetForm()
  }
}, { immediate: true })

// 表单数据
const form = ref({
  orgId: ''
})

// 文件相关
const uploadRef = ref<UploadInstance>()
const fileList = ref<UploadFile[]>([])
const selectedFile = ref<File | null>(null)

// 状态
const uploading = ref(false)
const importTask = ref<ImportTask | null>(null)
const pollingTimer = ref<number | null>(null)

// 计算属性
const canSubmit = computed(() => {
  return props.orgId && selectedFile.value && !uploading.value
})

const statusType = computed(() => {
  if (!importTask.value) return 'info'
  switch (importTask.value.status) {
    case 'success':
      return 'success'
    case 'failed':
      return 'danger'
    case 'processing':
      return 'warning'
    default:
      return 'info'
  }
})

const statusText = computed(() => {
  if (!importTask.value) return ''
  return importTask.value.message
})

const progressStatus = computed(() => {
  if (!importTask.value) return undefined
  if (importTask.value.status === 'success') return 'success'
  if (importTask.value.status === 'failed') return 'exception'
  return undefined
})

// 文件选择
const handleFileChange = (file: UploadFile) => {
  selectedFile.value = file.raw || null
}

const handleFileRemove = () => {
  selectedFile.value = null
}

// 重置表单
const resetForm = () => {
  fileList.value = []
  selectedFile.value = null
  importTask.value = null
  uploading.value = false
  stopPolling()
}

// 重置
const handleReset = () => {
  resetForm()
}

// 关闭提示框
const handleCloseAlert = () => {
  importTask.value = null
  stopPolling()
}

// 取消导入
const handleCancel = () => {
  if (uploading.value || importTask.value) {
    ElMessage.warning('导入过程中无法取消')
    return
  }
  stopPolling()
  emit('close')
}

// 提交导入
const handleSubmit = async () => {
  if (!canSubmit.value) return

  // 第一次尝试导入（不强制删除）
  await performImport(false)
}

// 执行导入
const performImport = async (forceDelete: boolean) => {
  try {
    uploading.value = true

    const formData = new FormData()
    formData.append('org_id', props.orgId)
    formData.append('file', selectedFile.value!)
    if (forceDelete) {
      formData.append('force_delete', 'true')
    }

    const response = await InquiryAPI.importInquiry(formData)
    const data = response.data

    // 成功：按你后端成功响应结构处理
    ElMessage.success(data.message || '导入任务已提交')
    startPolling(data.task_id)
  } catch (error: any) {

    const ax = error as AxiosError<ImportError>
    console.log('ax:', JSON.stringify(ax, null, 2))
    const payload = ax.response?.data

    console.log('payload:', JSON.stringify(payload, null, 2))

    const code = payload?.details?.code
    if (code === 1001 || code === 'DUPLICATE_INQUIRY') {
      const d = payload!.details!
      const duplicateType = d.type === 'title' ? '标题' : '日期'
      const msg = d.message || '发现重复的询价单'
      await ElMessageBox.confirm(
        `${msg}：已存在相同${duplicateType}的询价单（${d.value ?? ''}），是否强制导入？`,
        '检测到重复记录',
        { confirmButtonText: '强制导入', cancelButtonText: '取消', type: 'warning' }
      )
      return await performImport(true)
    }
  } finally {
    // 放在 finally，避免某些分支未复位
    uploading.value = false
  }
}

// 开始轮询
const startPolling = (taskId: string) => {
  stopPolling() // 先停止之前的轮询

  // 立即查询一次
  pollStatus(taskId)

  // 每2秒查询一次
  pollingTimer.value = window.setInterval(() => {
    pollStatus(taskId)
  }, 2000)
}

// 停止轮询
const stopPolling = () => {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value)
    pollingTimer.value = null
  }
}

// 查询状态
const pollStatus = async (taskId: string) => {
  try {
    const response = await InquiryAPI.getImportStatus(taskId)
    importTask.value = response.data
    uploading.value = false

    // 如果完成或失败，停止轮询
    if (importTask.value.status === 'success' || importTask.value.status === 'failed') {
      stopPolling()
      if (importTask.value.status === 'success') {
        emit('importSuccess')
      }
    }
  } catch (error: any) {
    console.error('查询导入状态失败:', error)
    stopPolling()
    uploading.value = false
  }
}

// 组件卸载时停止轮询
import { onUnmounted } from 'vue'
import type { AxiosError } from 'axios'
onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.inquiry-import {
  padding: 0;
}

.upload-section {
  padding: 0;
}

.tips p {
  margin: 4px 0;
  line-height: 1.6;
}

.alert-compact {
  width: fit-content;
  max-width: 100%;
}

.button-group {
  text-align: center;
}

.button-group .el-button {
  margin: 0 8px;
}

.progress-section {
  margin-top: 20px;
}

.task-info {
  margin-top: 16px;
}

.task-details p {
  margin: 8px 0;
  line-height: 1.6;
}

:deep(.el-upload-dragger) {
  padding: 40px;
}

:deep(.el-icon--upload) {
  font-size: 67px;
  color: #c0c4cc;
  margin-bottom: 16px;
}

:deep(.el-upload__text) {
  color: #606266;
  font-size: 14px;
  text-align: center;
}

:deep(.el-upload__text em) {
  color: #409eff;
  font-style: normal;
}

:deep(.el-upload__tip) {
  color: #909399;
  font-size: 12px;
  text-align: center;
  margin-top: 7px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .inquiry-import {
    padding: 10px;
  }
  
  .upload-section {
    padding: 10px 0;
  }
  
  :deep(.el-upload-dragger) {
    padding: 20px;
  }
}
</style>

