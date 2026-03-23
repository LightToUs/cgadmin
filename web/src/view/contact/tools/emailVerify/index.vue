<template>
  <div>
    <div class="gva-table-box">
      <div class="font-medium">批量邮箱验证</div>
      <el-divider />

      <div class="grid grid-cols-2 gap-4" style="max-width: 960px">
        <el-card shadow="never">
          <div class="text-sm font-medium mb-2">验证范围</div>
          <el-radio-group v-model="scopeType">
            <el-radio label="currentList">当前列表</el-radio>
            <el-radio label="allUnverified">全部未验证</el-radio>
          </el-radio-group>

          <div v-if="scopeType === 'currentList'" class="mt-3">
            <div class="text-sm text-gray-500 mb-2">选择列表</div>
            <el-select v-model="listId" class="w-full" placeholder="请选择列表">
              <el-option v-for="l in selectableLists" :key="l.id" :label="l.name" :value="l.id" />
            </el-select>
          </div>
        </el-card>

        <el-card shadow="never">
          <div class="text-sm font-medium mb-2">验证方式</div>
          <el-radio-group v-model="method">
            <el-radio label="basic">基础验证（免费）</el-radio>
            <el-radio label="deep">深度验证（消耗积分）</el-radio>
          </el-radio-group>

          <div class="mt-3 text-sm text-gray-700">
            预计消耗：{{ estimatedCost }} 积分
          </div>

          <div class="mt-4">
            <el-button type="primary" :loading="starting" @click="start">开始验证</el-button>
          </div>
        </el-card>
      </div>
    </div>

    <div class="gva-table-box mt-3">
      <div class="flex items-center justify-between">
        <div class="font-medium">验证进度</div>
        <el-button @click="refreshJob" :disabled="!jobId">刷新</el-button>
      </div>
      <el-divider />

      <el-empty v-if="!jobId" description="暂无任务" />

      <div v-else>
        <div class="flex items-center justify-between flex-wrap gap-2">
          <div class="text-sm text-gray-700">
            状态：<el-tag :type="jobStatusType(job.status)">{{ jobStatusText(job.status) }}</el-tag>
          </div>
          <div class="text-sm text-gray-700">进度：{{ job.progress }}%</div>
        </div>
        <el-progress class="mt-2" :percentage="job.progress || 0" />
        <div class="mt-3 text-sm text-gray-700">
          ✅ 有效{{ job.valid || 0 }} / ⚠️ 风险{{ job.risk || 0 }} / ❌ 无效{{ job.invalid || 0 }}
        </div>
      </div>
    </div>

    <div class="gva-table-box mt-3">
      <div class="flex items-center justify-between">
        <div class="font-medium">验证历史</div>
        <el-button @click="fetchHistory">刷新</el-button>
      </div>
      <el-divider />

      <el-empty v-if="!history.length" description="暂无记录" />
      <el-table v-else :data="history" style="width: 100%" row-key="id">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column label="状态" width="120">
          <template #default="scope">
            <el-tag :type="jobStatusType(scope.row.status)">{{ jobStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="范围" min-width="180">
          <template #default="scope">
            <span>{{ scope.row.scopeType }}</span>
          </template>
        </el-table-column>
        <el-table-column label="结果" min-width="240">
          <template #default="scope">
            <span>有效{{ scope.row.valid || 0 }} / 风险{{ scope.row.risk || 0 }} / 无效{{ scope.row.invalid || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="200">
          <template #default="scope">{{ formatDate(scope.row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button type="primary" link @click="setCurrent(scope.row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="historyTotal > 0" class="gva-pagination">
        <el-pagination
          :current-page="historyPage"
          :page-size="historyPageSize"
          :total="historyTotal"
          layout="total, prev, pager, next"
          @current-change="changeHistoryPage"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
  import { computed, onBeforeUnmount, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { formatDate } from '@/utils/format'
  import { getContactListTree } from '@/api/contactList'
  import { getEmailVerifyHistory, getEmailVerifyJob, startEmailVerify } from '@/api/emailVerify'

  defineOptions({ name: 'ContactEmailVerify' })

  const scopeType = ref('allUnverified')
  const method = ref('basic')
  const listId = ref(null)

  const tree = ref([])
  const loadTree = async () => {
    const res = await getContactListTree()
    if (res.code === 0) {
      tree.value = res.data.tree || []
    }
  }

  const flatten = (nodes) => {
    const out = []
    const walk = (arr) => {
      for (const n of arr || []) {
        out.push(n)
        if (n.children?.length) walk(n.children)
      }
    }
    walk(nodes)
    return out
  }

  const selectableLists = computed(() => {
    const all = flatten(tree.value)
    return all.filter(i => i.type === 'custom')
  })

  const estimatedCost = computed(() => {
    if (method.value === 'deep') return 0
    return 0
  })

  const jobId = ref(0)
  const job = ref({})
  const starting = ref(false)
  let timer = null

  const jobStatusType = (s) => {
    if (s === 'finished') return 'success'
    if (s === 'failed') return 'danger'
    if (s === 'running') return 'warning'
    return 'info'
  }
  const jobStatusText = (s) => {
    if (s === 'finished') return '完成'
    if (s === 'failed') return '失败'
    if (s === 'running') return '进行中'
    return s || '-'
  }

  const startPolling = () => {
    stopPolling()
    timer = setInterval(() => refreshJob(), 2000)
  }
  const stopPolling = () => {
    if (timer) clearInterval(timer)
    timer = null
  }
  onBeforeUnmount(() => stopPolling())

  const refreshJob = async () => {
    if (!jobId.value) return
    const res = await getEmailVerifyJob({ id: jobId.value })
    if (res.code === 0) {
      job.value = res.data.job || {}
      if (job.value.status === 'finished' || job.value.status === 'failed') {
        stopPolling()
      }
    }
  }

  const start = async () => {
    if (scopeType.value === 'currentList' && !listId.value) {
      ElMessage.warning('请选择列表')
      return
    }
    starting.value = true
    try {
      const res = await startEmailVerify({
        scopeType: scopeType.value,
        listId: scopeType.value === 'currentList' ? listId.value : null,
        method: method.value,
      })
      if (res.code === 0) {
        jobId.value = res.data.job.id
        job.value = res.data.job
        startPolling()
        fetchHistory()
      }
    } finally {
      starting.value = false
    }
  }

  const history = ref([])
  const historyPage = ref(1)
  const historyPageSize = ref(10)
  const historyTotal = ref(0)

  const fetchHistory = async () => {
    const res = await getEmailVerifyHistory({ page: historyPage.value, pageSize: historyPageSize.value })
    if (res.code === 0) {
      history.value = res.data.list || []
      historyTotal.value = res.data.total || 0
    }
  }
  const changeHistoryPage = (p) => {
    historyPage.value = p
    fetchHistory()
  }
  const setCurrent = (row) => {
    jobId.value = row.id
    refreshJob()
    startPolling()
  }

  loadTree().then(() => fetchHistory())
</script>
