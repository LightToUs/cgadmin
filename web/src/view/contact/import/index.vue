<template>
  <div>
    <div class="gva-table-box">
      <div class="flex items-center justify-between flex-wrap gap-2">
        <div class="font-medium">导入客户数据</div>
        <div class="flex items-center gap-2">
          <el-button @click="downloadTemplate('xlsx')">下载模板</el-button>
          <el-button @click="openGoogleSheet">从Google Sheets导入</el-button>
        </div>
      </div>
      <el-divider />

      <el-steps :active="step" finish-status="success" align-center>
        <el-step title="上传" />
        <el-step title="字段匹配" />
        <el-step title="数据验证" />
        <el-step title="导入结果" />
      </el-steps>
    </div>

    <div v-if="step === 0" class="gva-table-box">
      <div class="flex justify-center">
        <el-upload
          drag
          :auto-upload="false"
          :limit="1"
          :on-change="onFileChange"
          :on-remove="onFileRemove"
          accept=".csv,.xlsx,.xls"
        >
          <div class="py-8">
            <div class="text-lg">拖拽或点击上传</div>
            <div class="text-sm text-gray-500 mt-2">CSV / Excel，最大10MB</div>
          </div>
        </el-upload>
      </div>
      <div class="mt-4 flex justify-end">
        <el-button type="primary" :disabled="!fileRaw" :loading="uploading" @click="doUpload">下一步</el-button>
      </div>
    </div>

    <div v-else-if="step === 1" class="gva-table-box">
      <div class="flex items-center justify-between">
        <div class="text-gray-600 text-sm">系统字段与文件列匹配，标 * 为必填</div>
        <el-button @click="autoMatch">自动匹配</el-button>
      </div>
      <el-divider />
      <el-table :data="fieldRows" style="width: 100%">
        <el-table-column label="系统字段" min-width="180">
          <template #default="scope">
            <span>{{ scope.row.label }}</span>
            <span v-if="scope.row.required" class="text-red-500 ml-1">*</span>
          </template>
        </el-table-column>
        <el-table-column label="文件列" min-width="260">
          <template #default="scope">
            <el-select v-model="mapping[scope.row.key]" class="w-full" clearable placeholder="不导入">
              <el-option label="不导入" value="" />
              <el-option v-for="c in columns" :key="c" :label="c" :value="c" />
            </el-select>
          </template>
        </el-table-column>
      </el-table>
      <div class="mt-4 flex items-center justify-between">
        <el-button @click="step = 0">上一步</el-button>
        <el-button type="primary" :loading="validating" @click="goValidate">下一步</el-button>
      </div>
    </div>

    <div v-else-if="step === 2" class="gva-table-box">
      <div class="text-sm text-gray-700">
        <div>总记录：{{ validateStat.total }}</div>
        <div class="mt-2">✅ 格式正确：{{ validateStat.validCount }}</div>
        <div class="mt-2 flex items-center gap-2">
          <span>❌ 格式错误：{{ validateStat.invalidCount }}</span>
          <el-button v-if="validateStat.invalidCount" type="primary" link @click="openErrorDialog('invalid_email')">查看</el-button>
        </div>
        <div class="mt-2 flex items-center gap-2">
          <span>🔄 重复记录：{{ validateStat.duplicateCount }}</span>
          <el-button v-if="validateStat.duplicateCount" type="primary" link @click="openErrorDialog('duplicate')">查看</el-button>
        </div>
        <div class="mt-2 flex items-center gap-2">
          <span>⚠️ 缺少必填：{{ validateStat.missingRequiredCount }}</span>
          <el-button v-if="validateStat.missingRequiredCount" type="primary" link @click="openErrorDialog('missing_required')">查看</el-button>
        </div>
      </div>

      <el-divider />

      <div class="text-sm font-medium mb-2">处理选项</div>
      <el-radio-group v-model="options.onInvalid">
        <el-radio label="skip">跳过错误行，导入正确的</el-radio>
        <el-radio label="stop">停止导入，我手动修改</el-radio>
      </el-radio-group>

      <div class="mt-4 flex items-center justify-between">
        <el-button @click="step = 1">上一步</el-button>
        <el-button type="primary" :loading="starting" @click="startImport">开始导入</el-button>
      </div>
    </div>

    <div v-else class="gva-table-box">
      <div v-if="job.status === 'running' || job.status === 'validated' || job.status === 'uploaded'">
        <div class="text-sm text-gray-700 mb-2">
          正在导入：{{ job.progress }}%
        </div>
        <el-progress :percentage="job.progress || 0" />
        <div class="mt-3 text-sm text-gray-700">
          成功：{{ (job.createdCount || 0) + (job.updatedCount || 0) }} / 失败：{{ job.failedCount || 0 }}
        </div>
        <div class="mt-4 flex gap-2">
          <el-button @click="goBackground">后台运行</el-button>
          <el-button @click="pollJob">刷新</el-button>
        </div>
      </div>

      <div v-else-if="job.status === 'finished'">
        <div class="text-lg font-medium">✅ 导入完成！</div>
        <div class="mt-3 text-sm text-gray-700">
          <div>总处理：{{ job.total }}</div>
          <div>成功：{{ (job.createdCount || 0) + (job.updatedCount || 0) }}（新增{{ job.createdCount || 0 }} / 更新{{ job.updatedCount || 0 }}）</div>
          <div>失败：{{ job.failedCount || 0 }}</div>
        </div>
        <div class="mt-4 flex flex-wrap gap-2">
          <el-button type="primary" @click="goList">查看客户</el-button>
          <el-button @click="goVerify">批量验证邮箱</el-button>
          <el-button :disabled="!job.failedCount" @click="downloadFailed">导出失败记录</el-button>
        </div>
      </div>

      <div v-else>
        <el-alert type="error" show-icon :title="job.errorMessage || '导入失败'" />
        <div class="mt-4 flex gap-2">
          <el-button @click="step = 0">重新上传</el-button>
        </div>
      </div>
    </div>

    <el-dialog v-model="errorDialogVisible" title="问题数据" width="900px" destroy-on-close>
      <el-table :data="errorRows" style="width: 100%">
        <el-table-column prop="rowIndex" label="行号" width="90" />
        <el-table-column prop="errorsJson" label="错误" width="220" />
        <el-table-column prop="rawJson" label="原始数据" />
      </el-table>
      <div class="mt-3 flex justify-end">
        <el-pagination
          :current-page="errorPage"
          :page-size="errorPageSize"
          :total="errorTotal"
          layout="total, prev, pager, next"
          @current-change="changeErrorPage"
        />
      </div>
    </el-dialog>

    <el-dialog v-model="googleSheetVisible" title="从Google Sheets导入" width="560px" destroy-on-close>
      <el-input v-model.trim="googleSheetUrl" placeholder="请输入公开的CSV下载链接" />
      <div class="text-xs text-gray-500 mt-2">
        建议使用Google Sheets的“发布到网络”的CSV链接。
      </div>
      <template #footer>
        <el-button @click="googleSheetVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="doGoogleSheet">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, onBeforeUnmount, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRoute, useRouter } from 'vue-router'
  import {
    exportContactImportFailed,
    getContactImportErrors,
    getContactImportJob,
    startContactImport,
    suggestContactImportMapping,
    uploadContactImport,
    uploadContactImportGoogleSheet,
    validateContactImport
  } from '@/api/contactImport'

  defineOptions({ name: 'ContactImport' })

  const router = useRouter()
  const route = useRoute()

  const step = ref(0)
  const jobId = ref(0)
  const job = ref({})

  const fileRaw = ref(null)
  const uploading = ref(false)

  const columns = ref([])
  const mapping = ref({
    companyName: '',
    website: '',
    contactName: '',
    title: '',
    email: '',
    phone: '',
    country: '',
  })

  const options = ref({
    onInvalid: 'skip',
    onDuplicate: 'update',
    listId: null,
  })

  const fieldRows = computed(() => ([
    { key: 'companyName', label: '公司名称', required: true },
    { key: 'email', label: '邮箱', required: true },
    { key: 'contactName', label: '联系人', required: false },
    { key: 'title', label: '职位', required: false },
    { key: 'phone', label: '电话', required: false },
    { key: 'country', label: '国家', required: false },
    { key: 'website', label: '网站', required: false },
  ]))

  const onFileChange = (file) => {
    fileRaw.value = file?.raw || null
  }

  const onFileRemove = () => {
    fileRaw.value = null
  }

  const doUpload = async () => {
    if (!fileRaw.value) return
    uploading.value = true
    try {
      const formData = new FormData()
      formData.append('file', fileRaw.value)
      const res = await uploadContactImport(formData)
      if (res.code === 0) {
        jobId.value = res.data.jobId
        columns.value = res.data.columns || []
        step.value = 1
        await autoMatch()
      }
    } finally {
      uploading.value = false
    }
  }

  const autoMatch = async () => {
    if (!jobId.value) return
    const res = await suggestContactImportMapping({ jobId: jobId.value })
    if (res.code === 0) {
      columns.value = res.data.columns || columns.value
      const s = res.data.suggest || {}
      mapping.value = { ...mapping.value, ...s }
    }
  }

  const validating = ref(false)
  const validateStat = ref({
    total: 0,
    validCount: 0,
    invalidCount: 0,
    duplicateCount: 0,
    missingRequiredCount: 0,
  })

  const goValidate = async () => {
    if (!mapping.value.companyName || !mapping.value.email) {
      ElMessage.warning('请至少匹配公司名称与邮箱')
      return
    }
    validating.value = true
    try {
      const res = await validateContactImport({
        jobId: jobId.value,
        mapping: mapping.value,
        options: options.value,
      })
      if (res.code === 0) {
        validateStat.value = res.data
        step.value = 2
      }
    } finally {
      validating.value = false
    }
  }

  const starting = ref(false)
  let timer = null

  const startImport = async () => {
    starting.value = true
    try {
      const res = await startContactImport({
        jobId: jobId.value,
        mapping: mapping.value,
        options: options.value,
      })
      if (res.code === 0) {
        step.value = 3
        await pollJob()
        startPolling()
      }
    } finally {
      starting.value = false
    }
  }

  const pollJob = async () => {
    if (!jobId.value) return
    const res = await getContactImportJob({ id: jobId.value })
    if (res.code === 0) {
      job.value = res.data.job || {}
      if (job.value.status === 'finished' || job.value.status === 'failed') {
        stopPolling()
      }
    }
  }

  const startPolling = () => {
    stopPolling()
    timer = setInterval(() => {
      pollJob()
    }, 2000)
  }

  const stopPolling = () => {
    if (timer) clearInterval(timer)
    timer = null
  }

  onBeforeUnmount(() => stopPolling())

  const downloadTemplate = (format) => {
    let baseUrl = import.meta.env.VITE_BASE_API
    if (baseUrl === '/') baseUrl = ''
    window.open(`${baseUrl}/contactImport/template?format=${format}`, '_blank')
  }

  const goList = () => router.push({ name: 'contactMyList' })
  const goVerify = () => router.push({ name: 'contactEmailVerify' })
  const goBackground = () => router.push({ name: 'contactImportHistory' })

  const downloadFailed = async () => {
    const res = await exportContactImportFailed({ jobId: jobId.value })
    if (res.code === 0) {
      let baseUrl = import.meta.env.VITE_BASE_API
      if (baseUrl === '/') baseUrl = ''
      window.open(`${baseUrl}/contactImport/downloadByToken?token=${res.data.token}`, '_blank')
    }
  }

  const errorDialogVisible = ref(false)
  const errorType = ref('')
  const errorRows = ref([])
  const errorPage = ref(1)
  const errorPageSize = ref(20)
  const errorTotal = ref(0)

  const openErrorDialog = async (type) => {
    errorType.value = type
    errorPage.value = 1
    errorDialogVisible.value = true
    await fetchErrorRows()
  }

  const fetchErrorRows = async () => {
    const res = await getContactImportErrors({
      jobId: jobId.value,
      type: errorType.value,
      page: errorPage.value,
      pageSize: errorPageSize.value,
    })
    if (res.code === 0) {
      errorRows.value = res.data.list || []
      errorTotal.value = res.data.total || 0
    }
  }

  const changeErrorPage = async (p) => {
    errorPage.value = p
    await fetchErrorRows()
  }

  const googleSheetVisible = ref(false)
  const googleSheetUrl = ref('')
  const openGoogleSheet = () => {
    googleSheetUrl.value = ''
    googleSheetVisible.value = true
  }

  const doGoogleSheet = async () => {
    if (!googleSheetUrl.value) {
      ElMessage.warning('请输入链接')
      return
    }
    uploading.value = true
    try {
      const res = await uploadContactImportGoogleSheet({ url: googleSheetUrl.value })
      if (res.code === 0) {
        jobId.value = res.data.jobId
        columns.value = res.data.columns || []
        step.value = 1
        googleSheetVisible.value = false
        await autoMatch()
      }
    } finally {
      uploading.value = false
    }
  }

  const initFromJob = async (id) => {
    jobId.value = id
    const res = await getContactImportJob({ id })
    if (res.code === 0) {
      const j = res.data.job || {}
      job.value = j
      try {
        columns.value = JSON.parse(j.columnsJson || '[]')
      } catch {
        columns.value = []
      }
      try {
        const m = JSON.parse(j.mappingJson || '{}')
        mapping.value = { ...mapping.value, ...m }
      } catch {
      }
      step.value = j.status === 'finished' || j.status === 'failed' || j.status === 'running' ? 3 : 1
      if (step.value === 3 && (j.status === 'running' || j.status === 'validated' || j.status === 'uploaded')) {
        startPolling()
      }
    }
  }

  const qJobId = Number(route.query.jobId || 0)
  if (qJobId) {
    initFromJob(qJobId)
  }
</script>

