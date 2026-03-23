<template>
  <div class="gva-table-box">
    <div class="flex items-center justify-between">
      <div class="font-medium">导入历史</div>
      <el-button type="primary" @click="goImport">导入客户</el-button>
    </div>
    <el-divider />

    <el-empty v-if="!tableData.length" description="暂无导入记录" />

    <el-table v-else :data="tableData" style="width: 100%" row-key="id">
      <el-table-column label="时间" width="180">
        <template #default="scope">{{ formatDate(scope.row.createdAt) }}</template>
      </el-table-column>
      <el-table-column label="文件" prop="filename" min-width="220" />
      <el-table-column label="状态" width="120">
        <template #default="scope">
          <el-tag :type="statusType(scope.row.status)">{{ statusText(scope.row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="统计" min-width="260">
        <template #default="scope">
          <span v-if="scope.row.status === 'finished'">总{{ scope.row.total }}，成功{{ (scope.row.createdCount || 0) + (scope.row.updatedCount || 0) }}，失败{{ scope.row.failedCount || 0 }}</span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="320">
        <template #default="scope">
          <el-button type="primary" link @click="openDetail(scope.row)">查看</el-button>
          <el-button type="primary" link :disabled="!scope.row.failedCount" @click="downloadFailed(scope.row)">下载失败</el-button>
          <el-button type="primary" link @click="reimport(scope.row)">再次导入</el-button>
          <el-button type="danger" link @click="remove(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div v-if="total > 0" class="gva-pagination">
      <el-pagination
        :current-page="page"
        :page-size="pageSize"
        :page-sizes="[10, 30, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handleCurrentChange"
        @size-change="handleSizeChange"
      />
    </div>

    <el-dialog v-model="detailVisible" title="导入详情" width="720px" destroy-on-close>
      <div v-if="detail">
        <div class="text-sm text-gray-700">
          <div>文件：{{ detail.filename }}</div>
          <div class="mt-2">状态：{{ statusText(detail.status) }}</div>
          <div class="mt-2">进度：{{ detail.progress }}%</div>
          <div class="mt-2">总记录：{{ detail.total }}</div>
          <div class="mt-2">新增：{{ detail.createdCount || 0 }}，更新：{{ detail.updatedCount || 0 }}，失败：{{ detail.failedCount || 0 }}</div>
          <div v-if="detail.errorMessage" class="mt-2 text-red-500">{{ detail.errorMessage }}</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { deleteContactImportJob, exportContactImportFailed, getContactImportHistory, getContactImportJob } from '@/api/contactImport'
  import { formatDate } from '@/utils/format'

  defineOptions({ name: 'ContactImportHistory' })

  const router = useRouter()

  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const tableData = ref([])

  const statusType = (s) => {
    if (s === 'finished') return 'success'
    if (s === 'failed') return 'danger'
    if (s === 'running') return 'warning'
    return 'info'
  }

  const statusText = (s) => {
    if (s === 'finished') return '完成'
    if (s === 'failed') return '失败'
    if (s === 'running') return '进行中'
    if (s === 'validated') return '已验证'
    if (s === 'uploaded') return '已上传'
    return s || '-'
  }

  const fetchList = async () => {
    const res = await getContactImportHistory({ page: page.value, pageSize: pageSize.value })
    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    }
  }

  const handleCurrentChange = (p) => {
    page.value = p
    fetchList()
  }

  const handleSizeChange = (s) => {
    pageSize.value = s
    page.value = 1
    fetchList()
  }

  const detailVisible = ref(false)
  const detail = ref(null)
  const openDetail = async (row) => {
    const res = await getContactImportJob({ id: row.id })
    if (res.code === 0) {
      detail.value = res.data.job
      detailVisible.value = true
    }
  }

  const downloadFailed = async (row) => {
    const res = await exportContactImportFailed({ jobId: row.id })
    if (res.code === 0) {
      let baseUrl = import.meta.env.VITE_BASE_API
      if (baseUrl === '/') baseUrl = ''
      window.open(`${baseUrl}/contactImport/downloadByToken?token=${res.data.token}`, '_blank')
    }
  }

  const reimport = (row) => {
    router.push({ name: 'contactImport', query: { jobId: row.id } })
  }

  const remove = async (row) => {
    await ElMessageBox.confirm('确定要删除该导入记录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await deleteContactImportJob({ id: row.id })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchList()
    }
  }

  const goImport = () => router.push({ name: 'contactImport' })

  fetchList()
</script>

