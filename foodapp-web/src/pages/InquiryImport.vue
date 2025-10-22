<template>
  <div class="inquiry-import">
    <h2 style="margin: 8px 0 16px;">询价单导入</h2>

    <el-card>
      <div class="upload-section">
        <div class="tips">
          <el-alert
            title="导入说明"
            type="info"
            :closable="false"
            show-icon
          >
            <p>1. 请上传标准格式的Excel文件（.xlsx 或 .xls）</p>
            <p>2. 如果存在相同标题或日期的询价单，系统会提示是否覆盖</p>
            <p>3. 导入过程会在后台执行，您可以查看实时进度</p>
          </el-alert>
        </div>

        <el-form :model="form" label-width="120px" style="margin-top: 20px;">
          <el-form-item label="组织ID" required>
            <el-input 
              v-model="form.orgId" 
              placeholder="请输入组织ID"
              style="width: 400px"
            />
          </el-form-item>

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

          <el-form-item>
            <el-button 
              type="primary" 
              @click="handleSubmit"
              :loading="uploading"
              :disabled="!canSubmit"
            >
              {{ uploading ? '上传中...' : '开始导入' }}
            </el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 进度显示 -->
      <div v-if="importTask" class="progress-section">
        <el-divider />
        <h3>导入进度</h3>
        
        <div class="task-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="任务ID">{{ importTask.task_id }}</el-descriptions-item>
            <el-descriptions-item label="文件名">{{ importTask.file_name }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="statusType">{{ statusText }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="进度">
              {{ importTask.progress }}% ({{ importTask.processed_sheets }}/{{ importTask.total_sheets }} sheets)
            </el-descriptions-item>
          </el-descriptions>

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
            :closable="false"
            show-icon
            style="margin-top: 16px;"
          >
            <p v-if="importTask.inquiry_id">询价单ID: {{ importTask.inquiry_id }}</p>
          </el-alert>

          <!-- 失败信息 -->
          <el-alert
            v-if="importTask.status === 'failed'"
            title="导入失败"
            type="error"
            :closable="false"
            show-icon
            style="margin-top: 16px;"
          >
            <p>{{ importTask.error_message || '未知错误' }}</p>
          </el-alert>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile, UploadInstance } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { InquiryAPI, type ImportTask } from '@/api/inquiry'

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
  return form.value.orgId && selectedFile.value && !uploading.value
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

// 重置
const handleReset = () => {
  form.value.orgId = ''
  fileList.value = []
  selectedFile.value = null
  importTask.value = null
  stopPolling()
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
    formData.append('org_id', form.value.orgId)
    formData.append('file', selectedFile.value!)
    if (forceDelete) {
      formData.append('force_delete', 'true')
    }

    const response = await InquiryAPI.importInquiry(formData)
    const data = response.data

    ElMessage.success(data.message)

    // 开始轮询任务状态
    startPolling(data.task_id)

  } catch (error: any) {
    uploading.value = false
    
    // 检查是否是重复错误
    const errorMsg = error.message || error.toString()
    if (errorMsg.includes('已存在相同标题') || errorMsg.includes('已存在相同日期')) {
      // 提示用户是否覆盖
      ElMessageBox.confirm(
        `${errorMsg}，是否删除旧记录并导入新数据？`,
        '检测到重复记录',
        {
          confirmButtonText: '确定覆盖',
          cancelButtonText: '取消',
          type: 'warning',
        }
      ).then(() => {
        // 用户确认，强制删除并重新导入
        performImport(true)
      }).catch(() => {
        ElMessage.info('已取消导入')
      })
    } else {
      ElMessage.error('导入失败: ' + errorMsg)
    }
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
    }
  } catch (error: any) {
    console.error('查询导入状态失败:', error)
    stopPolling()
    uploading.value = false
  }
}

// 组件卸载时停止轮询
import { onUnmounted } from 'vue'
onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.inquiry-import {
  padding: 20px;
}

.upload-section {
  padding: 20px 0;
}

.tips p {
  margin: 4px 0;
  line-height: 1.6;
}

.progress-section {
  margin-top: 20px;
}

.task-info {
  margin-top: 16px;
}

:deep(.el-upload-dragger) {
  padding: 40px;
}

:deep(.el-icon--upload) {
  font-size: 67px;
  margin-bottom: 16px;
}
</style>

