<template>
  <div class="personal-dashboard">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>个人数据看板</h1>
        <p>实时查看个人工单处理情况和效率</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showExportDialog = true">
          <el-icon><Download /></el-icon>
          导出数据
        </el-button>
        <el-button @click="showConfigDialog = true">
          <el-icon><Setting /></el-icon>
          配置
        </el-button>
        <el-button @click="fetchData" :icon="'Refresh'" circle />
        <span class="update-time">最后更新: {{ lastUpdateTime }}</span>
      </div>
    </div>

    <!-- 筛选条件 -->
    <div class="filter-bar">
      <div class="filter-row">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          @change="handleDateChange"
        />
        <el-select v-model="filterStatus" placeholder="工单状态" clearable @change="fetchData">
          <el-option label="处理中" value="processing" />
          <el-option label="待复测" value="retesting" />
          <el-option label="已关闭" value="closed" />
        </el-select>
        <el-select v-model="filterPriority" placeholder="优先级" clearable @change="fetchData">
          <el-option label="低" value="low" />
          <el-option label="中" value="medium" />
          <el-option label="高" value="high" />
          <el-option label="紧急" value="critical" />
        </el-select>
      </div>
      <div class="filter-row">
        <el-radio-group v-model="timeDimension" @change="fetchData">
          <el-radio-button value="day">按天</el-radio-button>
          <el-radio-button value="week">按周</el-radio-button>
          <el-radio-button value="month">按月</el-radio-button>
        </el-radio-group>
        <el-button type="primary" @click="fetchData">查询</el-button>
      </div>
    </div>

    <!-- 数据卡片 -->
    <div class="stats-cards">
      <div class="stat-card pending" v-if="dashboardConfig.showPending !== false" @click="showTicketList('pending')">
        <div class="card-icon">
          <el-icon><Clock /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">{{ stats.pending_tickets }}</div>
          <div class="card-label">待处理工单</div>
        </div>
      </div>
      <div class="stat-card processed" v-if="dashboardConfig.showProcessed !== false" @click="showTicketList('processed')">
        <div class="card-icon">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">{{ stats.processed_tickets }}</div>
          <div class="card-label">已处理工单</div>
        </div>
      </div>
      <div class="stat-card time" v-if="dashboardConfig.showAvgTime !== false">
        <div class="card-icon">
          <el-icon><Timer /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">{{ formatDuration(stats.avg_process_time) }}</div>
          <div class="card-label">平均处理时长</div>
        </div>
      </div>
      <div class="stat-card timeout" v-if="dashboardConfig.showTimeout !== false" @click="showTicketList('timeout')">
        <div class="card-icon">
          <el-icon><Warning /></el-icon>
        </div>
        <div class="card-content">
          <div class="card-value danger">{{ stats.timeout_tickets }}</div>
          <div class="card-label">超时工单</div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-grid">
      <!-- 趋势图 -->
      <div class="chart-card trend-chart" v-if="dashboardConfig.showTrend !== false">
        <div class="chart-header">
          <h3>工单处理趋势</h3>
          <el-radio-group v-model="trendType" size="small">
            <el-radio-button value="processed">处理量</el-radio-button>
            <el-radio-button value="timeout">超时量</el-radio-button>
          </el-radio-group>
        </div>
        <div ref="trendChartRef" class="chart-container"></div>
      </div>

      <!-- 状态分布 -->
      <div class="chart-card" v-if="dashboardConfig.showStatus !== false">
        <div class="chart-header">
          <h3>工单状态分布</h3>
        </div>
        <div ref="statusChartRef" class="chart-container"></div>
      </div>

      <!-- 优先级分布 -->
      <div class="chart-card" v-if="dashboardConfig.showPriority !== false">
        <div class="chart-header">
          <h3>工单优先级分布</h3>
        </div>
        <div ref="priorityChartRef" class="chart-container"></div>
      </div>
    </div>

    <!-- 导出弹窗 -->
    <el-dialog
      v-model="showExportDialog"
      title="导出工单数据"
      width="500px"
      destroy-on-close
    >
      <el-form :model="exportForm" label-width="100px">
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="exportForm.dateRange"
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
          <el-select v-model="exportForm.status" placeholder="全部状态" clearable style="width: 100%">
            <el-option label="处理中" value="processing" />
            <el-option label="待复测" value="retesting" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="exportForm.priority" placeholder="全部优先级" clearable style="width: 100%">
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="critical" />
          </el-select>
        </el-form-item>
        <el-form-item label="导出方式">
          <el-radio-group v-model="exportForm.async">
            <el-radio :value="false">同步导出（数据量小）</el-radio>
            <el-radio :value="true">异步导出（数据量大）</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExportDialog = false">取消</el-button>
        <el-button type="primary" @click="handleExport" :loading="exporting">
          {{ exporting ? '导出中...' : '确认导出' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 导出进度弹窗 -->
    <el-dialog
      v-model="showProgressDialog"
      title="导出进度"
      width="400px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div class="progress-content">
        <el-progress 
          :percentage="exportProgress" 
          :status="exportStatus === 'completed' ? 'success' : exportStatus === 'failed' ? 'exception' : ''"
        />
        <p class="progress-text">{{ progressText }}</p>
      </div>
      <template #footer>
        <el-button 
          v-if="exportStatus === 'completed'" 
          type="primary" 
          @click="downloadExport"
        >
          下载文件
        </el-button>
        <el-button 
          v-if="exportStatus === 'completed' || exportStatus === 'failed'" 
          @click="closeProgressDialog"
        >
          关闭
        </el-button>
      </template>
    </el-dialog>

    <!-- 配置弹窗 -->
    <el-dialog
      v-model="showConfigDialog"
      title="看板配置"
      width="500px"
    >
      <el-form label-width="120px">
        <el-form-item label="待处理工单">
          <el-switch v-model="dashboardConfig.showPending" />
        </el-form-item>
        <el-form-item label="已处理工单">
          <el-switch v-model="dashboardConfig.showProcessed" />
        </el-form-item>
        <el-form-item label="平均处理时长">
          <el-switch v-model="dashboardConfig.showAvgTime" />
        </el-form-item>
        <el-form-item label="超时工单">
          <el-switch v-model="dashboardConfig.showTimeout" />
        </el-form-item>
        <el-divider>图表展示</el-divider>
        <el-form-item label="趋势图">
          <el-switch v-model="dashboardConfig.showTrend" />
        </el-form-item>
        <el-form-item label="状态分布">
          <el-switch v-model="dashboardConfig.showStatus" />
        </el-form-item>
        <el-form-item label="优先级分布">
          <el-switch v-model="dashboardConfig.showPriority" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存配置</el-button>
      </template>
    </el-dialog>

    <!-- 工单列表弹窗 -->
    <el-dialog
      v-model="showTicketListDialog"
      :title="ticketListTitle"
      width="900px"
      destroy-on-close
    >
      <el-table :data="ticketList" v-loading="ticketListLoading" max-height="400">
        <el-table-column prop="jira_key" label="工单编号" width="150">
          <template #default="{ row }">
            <a :href="row.jira_url" target="_blank" class="jira-link">{{ row.jira_key }}</a>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="priority" label="优先级" width="80">
          <template #default="{ row }">
            <el-tag :type="getPriorityTagType(row.priority)" size="small">
              {{ getPriorityText(row.priority) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_timeout" label="超时" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.is_timeout" type="danger" size="small">是</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
      </el-table>
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="ticketListPage"
          v-model:page-size="ticketListPageSize"
          :total="ticketListTotal"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchTicketList"
          @current-change="fetchTicketList"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'
import { 
  getPersonalDashboard, 
  exportTicketsExcel, 
  exportTicketsExcelAsync, 
  downloadExcelFile, 
  getExportTaskStatus,
  getUserConfig,
  updateUserConfig,
  getDashboardTickets
} from '@/api'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'

dayjs.extend(duration)

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)

// 数据
const stats = reactive({
  pending_tickets: 0,
  processed_tickets: 0,
  avg_process_time: 0,
  timeout_tickets: 0,
  trend_data: [],
  status_distribution: [],
  priority_distribution: [],
  type_distribution: []
})

const dateRange = ref([])
const lastUpdateTime = ref('')
const trendType = ref('processed')
const timeDimension = ref('day')
const filterStatus = ref('')
const filterPriority = ref('')

// 看板配置
const dashboardConfig = reactive({
  showPending: true,
  showProcessed: true,
  showAvgTime: true,
  showTimeout: true,
  showTrend: true,
  showStatus: true,
  showPriority: true
})

// 图表引用
const trendChartRef = ref(null)
const statusChartRef = ref(null)
const priorityChartRef = ref(null)
let trendChart = null
let statusChart = null
let priorityChart = null

// 导出相关
const showExportDialog = ref(false)
const exporting = ref(false)
const exportForm = reactive({
  dateRange: [],
  status: '',
  priority: '',
  async: false
})

// 导出进度相关
const showProgressDialog = ref(false)
const exportProgress = ref(0)
const exportStatus = ref('pending')
const progressText = ref('准备导出...')
const currentTaskId = ref(null)
const downloadFilename = ref('')

// 配置弹窗
const showConfigDialog = ref(false)

// 工单列表弹窗
const showTicketListDialog = ref(false)
const ticketListTitle = ref('')
const ticketList = ref([])
const ticketListLoading = ref(false)
const ticketListPage = ref(1)
const ticketListPageSize = ref(10)
const ticketListTotal = ref(0)
const currentListType = ref('')

// 获取数据
const fetchData = async () => {
  try {
    const params = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    if (filterStatus.value) {
      params.status = filterStatus.value
    }
    if (filterPriority.value) {
      params.priority = filterPriority.value
    }
    if (timeDimension.value) {
      params.time_dimension = timeDimension.value
    }
    
    const data = await getPersonalDashboard(params)
    Object.assign(stats, data)
    
    lastUpdateTime.value = dayjs().format('HH:mm:ss')
    
    // 更新图表
    await nextTick()
    renderCharts()
  } catch (error) {
    console.error('获取数据失败:', error)
    ElMessage.error('获取数据失败')
  }
}

// 获取用户配置
const fetchUserConfig = async () => {
  try {
    const res = await getUserConfig()
    if (res.dashboard_config) {
      const config = JSON.parse(res.dashboard_config)
      Object.assign(dashboardConfig, config)
    }
  } catch (error) {
    console.error('获取配置失败:', error)
  }
}

// 保存配置
const saveConfig = async () => {
  try {
    await updateUserConfig({
      dashboard_config: JSON.stringify(dashboardConfig)
    })
    ElMessage.success('配置保存成功')
    showConfigDialog.value = false
    
    // 等待DOM更新后重新渲染图表
    await nextTick()
    
    // 销毁现有图表实例,以便重新创建
    if (trendChart) {
      trendChart.dispose()
      trendChart = null
    }
    if (statusChart) {
      statusChart.dispose()
      statusChart = null
    }
    if (priorityChart) {
      priorityChart.dispose()
      priorityChart = null
    }
    
    // 重新渲染图表
    await nextTick()
    renderCharts()
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存配置失败')
  }
}

// 处理日期变化
const handleDateChange = () => {
  fetchData()
}

// 格式化时长
const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const dur = dayjs.duration(seconds, 'seconds')
  const days = Math.floor(dur.asDays())
  const hours = dur.hours()
  const minutes = dur.minutes()

  let result = ''
  if (days > 0) result += `${days}天`
  if (hours > 0) result += `${hours}小时`
  if (minutes > 0 && days === 0) result += `${minutes}分钟`
  if (!result) result = '不到1分钟'

  return result
}

// 渲染图表
const renderCharts = () => {
  if (dashboardConfig.showTrend) {
    renderTrendChart()
  }
  if (dashboardConfig.showStatus) {
    renderStatusChart()
  }
  if (dashboardConfig.showPriority) {
    renderPriorityChart()
  }
}

// 渲染趋势图
const renderTrendChart = () => {
  if (!trendChartRef.value) return
  
  if (!trendChart) {
    trendChart = echarts.init(trendChartRef.value)
  }

  const dates = stats.trend_data?.map(item => item.date) || []
  const values = stats.trend_data?.map(item => 
    trendType.value === 'processed' ? item.processed : item.timeout
  ) || []

  const option = {
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        const data = stats.trend_data.find(d => d.date === params[0].axisValue)
        if (data) {
          return `${params[0].axisValue}<br/>
            新增: ${data.count}<br/>
            处理: ${data.processed}<br/>
            超时: ${data.timeout}`
        }
        return params[0].axisValue
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: {
        rotate: 45
      }
    },
    yAxis: {
      type: 'value'
    },
    series: [{
      name: trendType.value === 'processed' ? '处理量' : '超时量',
      type: 'line',
      smooth: true,
      data: values,
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: trendType.value === 'processed' ? 'rgba(102, 126, 234, 0.5)' : 'rgba(245, 101, 101, 0.5)' },
          { offset: 1, color: trendType.value === 'processed' ? 'rgba(102, 126, 234, 0.1)' : 'rgba(245, 101, 101, 0.1)' }
        ])
      },
      lineStyle: {
        color: trendType.value === 'processed' ? '#667eea' : '#f56565'
      },
      itemStyle: {
        color: trendType.value === 'processed' ? '#667eea' : '#f56565'
      }
    }]
  }

  trendChart.setOption(option)
  
  // 点击事件
  trendChart.off('click')
  trendChart.on('click', (params) => {
    const date = stats.trend_data[params.dataIndex]?.date
    if (date) {
      showTicketListByDate(date)
    }
  })
}

// 渲染状态分布图
const renderStatusChart = () => {
  if (!statusChartRef.value) return
  
  if (!statusChart) {
    statusChart = echarts.init(statusChartRef.value)
  }

  const data = stats.status_distribution?.map(item => ({
    name: item.status,
    value: item.count
  })) || []

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      bottom: '5%',
      left: 'center'
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: false
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 14,
          fontWeight: 'bold'
        }
      },
      labelLine: {
        show: false
      },
      data: data,
      color: ['#f56565', '#ed8936', '#48bb78']
    }]
  }

  statusChart.setOption(option)
  
  // 点击事件
  statusChart.off('click')
  statusChart.on('click', (params) => {
    showTicketListByStatus(params.name)
  })
}

// 渲染优先级分布图
const renderPriorityChart = () => {
  if (!priorityChartRef.value) return
  
  if (!priorityChart) {
    priorityChart = echarts.init(priorityChartRef.value)
  }

  const data = stats.priority_distribution?.map(item => ({
    name: item.priority,
    value: item.count
  })) || []

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      bottom: '5%',
      left: 'center'
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: false
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 14,
          fontWeight: 'bold'
        }
      },
      labelLine: {
        show: false
      },
      data: data,
      color: ['#48bb78', '#4299e1', '#ed8936', '#f56565']
    }]
  }

  priorityChart.setOption(option)

  // 点击事件
  priorityChart.off('click')
  priorityChart.on('click', (params) => {
    showTicketListByPriority(params.name)
  })
}

// 显示工单列表
const showTicketList = (type) => {
  currentListType.value = type
  ticketListPage.value = 1
  
  const titleMap = {
    pending: '待处理工单',
    processed: '已处理工单',
    timeout: '超时工单'
  }
  ticketListTitle.value = titleMap[type] || '工单列表'
  
  fetchTicketList()
  showTicketListDialog.value = true
}

// 按日期显示工单列表
const showTicketListByDate = (date) => {
  currentListType.value = 'date'
  ticketListTitle.value = `${date} 的工单`
  ticketListPage.value = 1
  fetchTicketList(date)
  showTicketListDialog.value = true
}

// 按状态显示工单列表
const showTicketListByStatus = (status) => {
  currentListType.value = 'status'
  ticketListTitle.value = `状态: ${status}`
  ticketListPage.value = 1
  fetchTicketList('', status)
  showTicketListDialog.value = true
}

// 按优先级显示工单列表
const showTicketListByPriority = (priority) => {
  currentListType.value = 'priority'
  ticketListTitle.value = `优先级: ${priority}`
  ticketListPage.value = 1
  fetchTicketList('', '', priority)
  showTicketListDialog.value = true
}

// 获取工单列表
const fetchTicketList = async (date = '', status = '', priority = '', ticketType = '') => {
  ticketListLoading.value = true
  try {
    const params = {
      page: ticketListPage.value,
      page_size: ticketListPageSize.value
    }
    
    if (date) params.date = date
    if (status) params.status = status
    if (priority) params.priority = priority
    if (ticketType) params.ticket_type = ticketType
    
    // 根据类型设置筛选条件
    if (currentListType.value === 'pending') {
      params.status = 'processing,retesting'
    } else if (currentListType.value === 'timeout') {
      // 超时工单需要特殊处理
    }
    
    const res = await getDashboardTickets(params)
    ticketList.value = res.list || []
    ticketListTotal.value = res.total || 0
  } catch (error) {
    console.error('获取工单列表失败:', error)
    ElMessage.error('获取工单列表失败')
  } finally {
    ticketListLoading.value = false
  }
}

// 导出数据
const handleExport = async () => {
  if (exportForm.async) {
    await handleAsyncExport()
  } else {
    await handleSyncExport()
  }
}

// 同步导出
const handleSyncExport = async () => {
  exporting.value = true
  try {
    const params = {}
    if (exportForm.dateRange && exportForm.dateRange.length === 2) {
      params.start_date = exportForm.dateRange[0]
      params.end_date = exportForm.dateRange[1]
    }
    if (exportForm.status) {
      params.status = exportForm.status
    }
    if (exportForm.priority) {
      params.priority = exportForm.priority
    }

    const result = await exportTicketsExcel(params)
    
    if (result.filename) {
      const blob = await downloadExcelFile(result.filename)
      
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = result.filename
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      
      ElMessage.success('导出成功')
      showExportDialog.value = false
    }
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

// 异步导出
const handleAsyncExport = async () => {
  exporting.value = true
  try {
    const params = {}
    if (exportForm.dateRange && exportForm.dateRange.length === 2) {
      params.start_date = exportForm.dateRange[0]
      params.end_date = exportForm.dateRange[1]
    }
    if (exportForm.status) {
      params.status = exportForm.status
    }
    if (exportForm.priority) {
      params.priority = exportForm.priority
    }

    const result = await exportTicketsExcelAsync(params)
    
    if (result.task_id) {
      currentTaskId.value = result.task_id
      showExportDialog.value = false
      showProgressDialog.value = true
      exportStatus.value = 'pending'
      exportProgress.value = 0
      progressText.value = '正在准备导出...'
      
      // 开始轮询任务状态
      pollExportStatus()
    }
  } catch (error) {
    console.error('创建导出任务失败:', error)
    ElMessage.error('创建导出任务失败')
  } finally {
    exporting.value = false
  }
}

// 轮询导出状态
const pollExportStatus = async () => {
  if (!currentTaskId.value) return
  
  try {
    const result = await getExportTaskStatus(currentTaskId.value)
    
    exportStatus.value = result.status
    
    if (result.status === 'processing') {
      exportProgress.value = 50
      progressText.value = '正在生成Excel文件...'
      setTimeout(pollExportStatus, 2000)
    } else if (result.status === 'completed') {
      exportProgress.value = 100
      progressText.value = '导出完成！'
      downloadFilename.value = result.filename
    } else if (result.status === 'failed') {
      exportProgress.value = 0
      progressText.value = `导出失败: ${result.error || '未知错误'}`
    }
  } catch (error) {
    console.error('获取任务状态失败:', error)
    progressText.value = '获取任务状态失败'
  }
}

// 下载导出文件
const downloadExport = async () => {
  if (!downloadFilename.value) return
  
  try {
    const blob = await downloadExcelFile(downloadFilename.value)
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = downloadFilename.value
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (error) {
    console.error('下载失败:', error)
    ElMessage.error('下载失败')
  }
}

// 关闭进度弹窗
const closeProgressDialog = () => {
  showProgressDialog.value = false
  currentTaskId.value = null
  downloadFilename.value = ''
}

// 辅助函数
const getPriorityText = (priority) => {
  const texts = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '紧急'
  }
  return texts[priority] || priority
}

const getPriorityTagType = (priority) => {
  const types = {
    low: 'info',
    medium: 'primary',
    high: 'warning',
    critical: 'danger'
  }
  return types[priority] || 'info'
}

const getStatusText = (status) => {
  const texts = {
    processing: '处理中',
    retesting: '待复测',
    closed: '已关闭'
  }
  return texts[status] || status
}

const getStatusTagType = (status) => {
  const types = {
    processing: 'warning',
    retesting: 'primary',
    closed: 'success'
  }
  return types[status] || 'info'
}

// 监听趋势类型变化
watch(trendType, () => {
  renderTrendChart()
})

// 窗口大小变化时重绘图表
const handleResize = () => {
  trendChart?.resize()
  statusChart?.resize()
  priorityChart?.resize()
}

// 定时刷新
let timer = null

onMounted(() => {
  fetchUserConfig()
  fetchData()
  timer = setInterval(fetchData, 60000) // 每分钟刷新
  
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
  
  trendChart?.dispose()
  statusChart?.dispose()
  priorityChart?.dispose()
  
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.personal-dashboard {
  padding: 0;
}

/* 页面标题 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left h1 {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.header-left p {
  color: var(--text-tertiary);
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.update-time {
  color: var(--text-tertiary);
  font-size: 13px;
}

/* 筛选条 */
.filter-bar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
  padding: 16px 20px;
  background: var(--bg-primary);
  border-radius: 12px;
  box-shadow: var(--shadow-card);
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: nowrap;
}

.filter-row > * {
  flex-shrink: 0;
}

/* 限制筛选组件宽度 */
.filter-row .el-date-editor--daterange {
  max-width: 280px;
}

.filter-row .el-select {
  max-width: 140px;
}

.filter-row .el-radio-group {
  max-width: 240px;
}

/* 数据卡片 */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
  transition: all var(--transition-base);
  cursor: pointer;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.card-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: white;
}

.stat-card.pending .card-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-card.processed .card-icon {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
}

.stat-card.time .card-icon {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-card.timeout .card-icon {
  background: linear-gradient(135deg, #f56565 0%, #e53e3e 100%);
}

.card-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.card-value.danger {
  color: var(--danger-color);
}

.card-label {
  color: var(--text-secondary);
  font-size: 14px;
  margin-top: 4px;
}

/* 图表区域 */
.charts-grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 20px;
  margin-bottom: 24px;
}

.chart-card {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 20px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
}

.chart-card.trend-chart {
  grid-column: span 1;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.chart-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.chart-container {
  height: 300px;
}

/* 进度弹窗 */
.progress-content {
  padding: 20px 0;
  text-align: center;
}

.progress-text {
  margin-top: 16px;
  color: var(--text-secondary);
}

/* 分页 */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

/* Jira链接 */
.jira-link {
  color: var(--primary-color);
  text-decoration: none;
}

.jira-link:hover {
  text-decoration: underline;
}

/* 响应式 */
@media (max-width: 1400px) {
  .charts-grid {
    grid-template-columns: 1fr 1fr;
  }
  
  .chart-card.trend-chart {
    grid-column: span 2;
  }
}

@media (max-width: 1200px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }
  
  .chart-card.trend-chart {
    grid-column: span 1;
  }
}

@media (max-width: 768px) {
  .stats-cards {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .header-right {
    width: 100%;
    justify-content: flex-end;
    flex-wrap: wrap;
  }
  
  .filter-bar {
    flex-direction: column;
  }
  
  .filter-row {
    flex-wrap: wrap;
  }
  
  .filter-bar .el-select,
  .filter-bar .el-date-picker {
    width: 100%;
  }
}
</style>
