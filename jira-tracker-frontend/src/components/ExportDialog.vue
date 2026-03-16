<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    title="导出工单数据"
    width="500px"
    destroy-on-close
  >
    <el-form :model="form" label-width="100px">
      <el-form-item label="时间范围">
        <el-date-picker
          v-model="form.dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          style="width: 100%"
        />
      </el-form-item>
      <el-form-item label="工单状态">
        <el-select v-model="form.status" placeholder="全部状态" clearable style="width: 100%">
          <el-option label="处理中" value="processing" />
          <el-option label="待复测" value="retesting" />
          <el-option label="已关闭" value="closed" />
        </el-select>
      </el-form-item>
      <el-form-item label="优先级">
        <el-select v-model="form.priority" placeholder="全部优先级" clearable style="width: 100%">
          <el-option label="低" value="low" />
          <el-option label="中" value="medium" />
          <el-option label="高" value="high" />
          <el-option label="紧急" value="critical" />
        </el-select>
      </el-form-item>
    </el-form>
    
    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="handleExport" :loading="loading">
        {{ loading ? '导出中...' : '确认导出' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { exportTicketsExcel, downloadExcelFile } from '@/api'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:visible', 'success'])

const loading = ref(false)
const form = reactive({
  dateRange: [],
  status: '',
  priority: ''
})

const handleExport = async () => {
  loading.value = true
  try {
    const params = {}
    
    if (form.dateRange && form.dateRange.length === 2) {
      params.start_date = form.dateRange[0]
      params.end_date = form.dateRange[1]
    }
    
    if (form.status) {
      params.status = form.status
    }
    
    if (form.priority) {
      params.priority = form.priority
    }

    const result = await exportTicketsExcel(params)
    
    if (result.filename) {
      // 使用axios下载文件
      const blob = await downloadExcelFile(result.filename)
      
      // 创建下载链接
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = result.filename
      document.body.appendChild(link)
      link.click()
      
      // 清理
      window.URL.revokeObjectURL(url)
      document.body.removeChild(link)
      
      ElMessage.success('导出成功')
      emit('update:visible', false)
      emit('success')
    }
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  } finally {
    loading.value = false
  }
}
</script>
