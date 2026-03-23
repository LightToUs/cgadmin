<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list flex justify-end">
        <el-button type="primary" icon="plus" @click="openCreate">+添加邮箱</el-button>
      </div>

      <el-empty
        v-if="!tableData.length"
        description="你还没有添加发件邮箱，添加一个开始营销吧"
        :image="emptyImage"
      >
        <el-button type="primary" icon="plus" @click="openCreate">+添加邮箱</el-button>
      </el-empty>

      <el-table
        v-else
        :data="tableData"
        style="width: 100%"
        row-key="id"
      >
        <el-table-column align="left" label="邮箱地址" prop="email" min-width="220" />
        <el-table-column align="left" label="显示名称" prop="displayName" min-width="140" />
        <el-table-column align="left" label="状态" width="140">
          <template #default="scope">
            <el-tooltip v-if="scope.row.status === 'error' && scope.row.lastError" :content="scope.row.lastError" placement="top">
              <el-tag :type="statusTagType(scope.row.status)">
                {{ statusText(scope.row.status) }}
              </el-tag>
            </el-tooltip>
            <el-tag v-else :type="statusTagType(scope.row.status)">
              {{ statusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="今日已发" width="160">
          <template #default="scope">
            <div class="flex items-center gap-2">
              <span>{{ scope.row.todaySent }}/{{ scope.row.dailyLimit }}</span>
              <el-progress
                :percentage="sentPercentage(scope.row)"
                :stroke-width="10"
                :show-text="false"
                :status="scope.row.status === 'quota_warning' ? 'warning' : undefined"
                style="width: 80px"
              />
            </div>
          </template>
        </el-table-column>
        <el-table-column align="left" label="默认" width="90">
          <template #default="scope">
            <el-tag v-if="scope.row.isDefault" type="success">默认</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="操作" min-width="340">
          <template #default="scope">
            <el-button
              type="primary"
              link
              :disabled="scope.row.isDefault"
              @click="setDefault(scope.row)"
            >设为默认</el-button>
            <el-button type="primary" link @click="openEdit(scope.row)">编辑</el-button>
            <el-button type="primary" link @click="openQuota(scope.row)">限额设置</el-button>
            <el-button
              type="primary"
              link
              :loading="scope.row.__testing"
              @click="testById(scope.row)"
            >测试连接</el-button>
            <el-button type="danger" link @click="removeRow(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-drawer
      v-model="drawerVisible"
      :title="drawerTitle"
      size="560px"
      destroy-on-close
    >
      <div class="p-2">
        <el-steps :active="activeStep" finish-status="success" align-center>
          <el-step title="选择类型" />
          <el-step title="填写配置" />
          <el-step title="测试连接" />
          <el-step title="保存" />
        </el-steps>

        <div class="mt-4">
          <div v-if="activeStep === 0" class="grid grid-cols-2 gap-3">
            <el-card
              v-for="item in providerOptions"
              :key="item.value"
              shadow="hover"
              class="cursor-pointer"
              :class="provider === item.value ? 'border border-solid border-[var(--el-color-primary)]' : ''"
              @click="selectProvider(item.value)"
            >
              <div class="flex items-center justify-between">
                <div class="font-medium">{{ item.label }}</div>
                <el-tag v-if="provider === item.value" type="success">已选</el-tag>
              </div>
              <div class="text-sm text-gray-500 mt-1">{{ item.desc }}</div>
            </el-card>
          </div>

          <div v-else-if="activeStep === 1">
            <el-form ref="configFormRef" :model="form" :rules="rules" label-width="120px">
              <el-form-item label="邮箱地址" prop="email">
                <el-input v-model.trim="form.email" placeholder="例如 mybusiness@qq.com" />
              </el-form-item>
              <el-form-item label="显示名称" prop="displayName">
                <el-input v-model.trim="form.displayName" placeholder="例如 Lisa Chen（可不填）" />
              </el-form-item>
              <el-form-item label="密码/授权码" prop="secret">
                <el-input v-model="form.secret" type="password" show-password placeholder="建议使用授权码" />
              </el-form-item>
              <el-form-item label="SMTP服务器" prop="smtpHost">
                <el-input v-model.trim="form.smtpHost" placeholder="例如 smtp.qq.com" />
              </el-form-item>
              <el-form-item label="端口" prop="smtpPort">
                <el-input-number v-model="form.smtpPort" :min="1" :max="65535" class="w-full" />
              </el-form-item>
              <el-form-item label="加密方式" prop="security">
                <el-select v-model="form.security" class="w-full">
                  <el-option label="SSL" value="ssl" />
                  <el-option label="TLS/STARTTLS" value="tls" />
                  <el-option label="无" value="none" />
                </el-select>
              </el-form-item>

              <el-divider />

              <el-form-item label="每日最大发送量" prop="dailyLimit">
                <el-input-number v-model="form.dailyLimit" :min="20" :max="200" class="w-full" />
              </el-form-item>
              <el-form-item label="发送间隔" prop="intervalPolicy">
                <el-select v-model="form.intervalPolicy" class="w-full">
                  <el-option label="1分钟" value="1m" />
                  <el-option label="3分钟" value="3m" />
                  <el-option label="5分钟" value="5m" />
                  <el-option label="10分钟" value="10m" />
                  <el-option label="随机间隔" value="random" />
                </el-select>
              </el-form-item>
            </el-form>

            <el-alert type="warning" show-icon class="mt-3">
              <template #title>
                <div class="font-medium">安全提示：为保护你的账号，建议使用“授权码”而非登录密码。</div>
              </template>
              <div class="mt-2">
                <el-collapse v-model="helpOpen">
                  <el-collapse-item title="QQ邮箱：如何获取授权码？" name="qq">
                    <div class="text-sm leading-6">
                      <div>1）登录QQ邮箱 → 设置 → 账户</div>
                      <div>2）开启 SMTP 服务</div>
                      <div>3）生成授权码并复制到这里</div>
                    </div>
                  </el-collapse-item>
                  <el-collapse-item title="163邮箱：如何获取授权码？" name="163">
                    <div class="text-sm leading-6">
                      <div>1）登录163邮箱 → 设置 → POP3/SMTP/IMAP</div>
                      <div>2）开启SMTP服务并生成授权码</div>
                    </div>
                  </el-collapse-item>
                  <el-collapse-item title="Gmail：如何开启应用专用密码？" name="gmail">
                    <div class="text-sm leading-6">
                      <div>1）Google账号安全设置 → 两步验证</div>
                      <div>2）启用“应用专用密码”并生成密码</div>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
            </el-alert>
          </div>

          <div v-else-if="activeStep === 2">
            <el-alert
              v-if="testState.status === 'success'"
              type="success"
              show-icon
              :title="testState.message"
            />
            <el-alert
              v-else-if="testState.status === 'error'"
              type="error"
              show-icon
              :title="testState.message"
            />
            <el-alert
              v-else
              type="info"
              show-icon
              title="点击“测试连接”，系统将发送一封测试邮件到该邮箱。"
            />

            <div class="mt-4 flex gap-3">
              <el-button
                type="primary"
                :loading="testState.loading"
                :disabled="testState.loading"
                @click="testConnection"
              >{{ testState.loading ? '测试中...' : '测试连接' }}</el-button>
              <el-button @click="activeStep = 1">返回修改</el-button>
            </div>
          </div>

          <div v-else>
            <el-alert type="success" show-icon title="测试通过，可以保存该邮箱账号。" />
            <div class="mt-4">
              <div class="text-sm text-gray-600">邮箱：{{ form.email }}</div>
              <div class="text-sm text-gray-600 mt-1">显示名称：{{ form.displayName || '-' }}</div>
              <div class="text-sm text-gray-600 mt-1">SMTP：{{ form.smtpHost }}:{{ form.smtpPort }}（{{ form.security.toUpperCase() }}）</div>
              <div class="text-sm text-gray-600 mt-1">安全限额：{{ form.todaySent || 0 }}/{{ form.dailyLimit }}，间隔策略：{{ intervalText(form.intervalPolicy) }}</div>
            </div>
            <div class="mt-4">
              <el-button type="primary" :loading="saving" @click="saveAccount">保存</el-button>
              <el-button @click="drawerVisible = false">取消</el-button>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-between">
          <el-button :disabled="activeStep === 0" @click="prevStep">上一步</el-button>
          <div class="flex gap-2">
            <el-button v-if="activeStep < 2" type="primary" @click="nextStep">下一步</el-button>
            <el-button v-else-if="activeStep === 2" type="primary" :disabled="testState.status !== 'success'" @click="activeStep = 3">下一步</el-button>
          </div>
        </div>
      </template>
    </el-drawer>

    <el-dialog v-model="quotaDialogVisible" title="安全限额设置" width="520px" destroy-on-close>
      <el-form :model="quotaForm" label-width="120px">
        <el-form-item label="每日最大发送量">
          <el-input-number v-model="quotaForm.dailyLimit" :min="20" :max="200" class="w-full" />
        </el-form-item>
        <el-form-item label="发送间隔">
          <el-select v-model="quotaForm.intervalPolicy" class="w-full">
            <el-option label="1分钟" value="1m" />
            <el-option label="3分钟" value="3m" />
            <el-option label="5分钟" value="5m" />
            <el-option label="10分钟" value="10m" />
            <el-option label="随机间隔" value="random" />
          </el-select>
        </el-form-item>
        <el-alert type="info" show-icon title="温馨提示：不同邮箱服务商官方限制不同，建议使用较保守的安全限额。" />
        <div class="text-sm text-gray-600 mt-2 leading-6">
          <div>QQ企业邮每天500封，个人QQ每天100封较安全</div>
          <div>Gmail、Outlook建议使用应用专用密码，配合更长发送间隔</div>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="quotaDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="quotaSaving" @click="saveQuota">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { computed, nextTick, reactive, ref, watch } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    createSenderEmailAccount,
    deleteSenderEmailAccount,
    getSenderEmailAccountList,
    setDefaultSenderEmailAccount,
    testSenderEmailAccount,
    testSenderEmailAccountById,
    updateSenderEmailAccount,
    updateSenderEmailAccountQuota,
  } from '@/api/senderEmailAccount'

  defineOptions({
    name: 'SendEmailAccount'
  })

  import emptyImg from '@/assets/noBody.png'

  const emptyImage = computed(() => emptyImg)

  const tableData = ref([])

  const fetchList = async () => {
    const res = await getSenderEmailAccountList()
    if (res.code === 0) {
      tableData.value = res.data.list || []
    }
  }

  const statusText = (status) => {
    if (status === 'normal') return '正常'
    if (status === 'error') return '异常'
    if (status === 'quota_warning') return '限额预警'
    return status || '-'
  }

  const statusTagType = (status) => {
    if (status === 'normal') return 'success'
    if (status === 'error') return 'danger'
    if (status === 'quota_warning') return 'warning'
    return 'info'
  }

  const sentPercentage = (row) => {
    if (!row || !row.dailyLimit) return 0
    const val = Math.floor((row.todaySent / row.dailyLimit) * 100)
    return Math.max(0, Math.min(100, val))
  }

  const intervalText = (policy) => {
    if (policy === '1m') return '1分钟'
    if (policy === '3m') return '3分钟'
    if (policy === '5m') return '5分钟'
    if (policy === '10m') return '10分钟'
    if (policy === 'random') return '随机间隔'
    return policy || '-'
  }

  const drawerVisible = ref(false)
  const drawerTitle = computed(() => (mode.value === 'create' ? '添加邮箱' : '编辑邮箱'))
  const activeStep = ref(0)
  const provider = ref('')
  const mode = ref('create')
  const editingId = ref(0)
  const configFormRef = ref(null)
  const helpOpen = ref([])

  const providerOptions = [
    { value: 'qq', label: 'QQ邮箱', desc: 'smtp.qq.com' },
    { value: '163', label: '163邮箱', desc: 'smtp.163.com' },
    { value: 'gmail', label: 'Gmail', desc: 'smtp.gmail.com' },
    { value: 'outlook', label: 'Outlook/Hotmail', desc: 'smtp.office365.com' },
    { value: 'custom', label: '企业邮/自定义SMTP', desc: '自定义服务器地址' },
  ]

  const defaultsByProvider = (p) => {
    if (p === 'qq') return { smtpHost: 'smtp.qq.com', smtpPort: 465, security: 'ssl' }
    if (p === '163') return { smtpHost: 'smtp.163.com', smtpPort: 465, security: 'ssl' }
    if (p === 'gmail') return { smtpHost: 'smtp.gmail.com', smtpPort: 587, security: 'tls' }
    if (p === 'outlook') return { smtpHost: 'smtp.office365.com', smtpPort: 587, security: 'tls' }
    return { smtpHost: '', smtpPort: 465, security: 'ssl' }
  }

  const form = reactive({
    email: '',
    displayName: '',
    secret: '',
    smtpHost: '',
    smtpPort: 465,
    security: 'ssl',
    isLoginAuth: false,
    dailyLimit: 50,
    intervalPolicy: '1m',
  })

  const rules = {
    email: [
      { required: true, message: '请输入邮箱地址', trigger: 'blur' },
      { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
    ],
    secret: [{ required: true, message: '请输入密码/授权码', trigger: 'blur' }],
    smtpHost: [{ required: true, message: '请输入SMTP服务器', trigger: 'blur' }],
    smtpPort: [{ required: true, message: '请输入端口', trigger: 'change' }],
    security: [{ required: true, message: '请选择加密方式', trigger: 'change' }],
    dailyLimit: [{ required: true, message: '请输入每日最大发送量', trigger: 'change' }],
    intervalPolicy: [{ required: true, message: '请选择发送间隔', trigger: 'change' }],
  }

  const testState = reactive({
    loading: false,
    status: '',
    message: ''
  })

  const saving = ref(false)

  watch(
    () => form.smtpPort,
    (val) => {
      if (val === 465) form.security = 'ssl'
      else if (val === 587) form.security = 'tls'
    }
  )

  const resetForm = () => {
    provider.value = ''
    activeStep.value = 0
    editingId.value = 0
    Object.assign(form, {
      email: '',
      displayName: '',
      secret: '',
      smtpHost: '',
      smtpPort: 465,
      security: 'ssl',
      isLoginAuth: false,
      dailyLimit: 50,
      intervalPolicy: '1m',
    })
    testState.loading = false
    testState.status = ''
    testState.message = ''
    helpOpen.value = []
  }

  const openCreate = async () => {
    mode.value = 'create'
    resetForm()
    drawerVisible.value = true
    await nextTick()
  }

  const openEdit = async (row) => {
    mode.value = 'edit'
    resetForm()
    drawerVisible.value = true
    provider.value = row.provider || 'custom'
    editingId.value = row.id
    activeStep.value = 1
    Object.assign(form, {
      email: row.email,
      displayName: row.displayName,
      secret: '',
      smtpHost: row.smtpHost,
      smtpPort: row.smtpPort,
      security: row.security,
      isLoginAuth: row.isLoginAuth,
      dailyLimit: row.dailyLimit,
      intervalPolicy: row.intervalPolicy,
    })
    await nextTick()
  }

  const selectProvider = (val) => {
    provider.value = val
    const d = defaultsByProvider(val)
    form.smtpHost = d.smtpHost
    form.smtpPort = d.smtpPort
    form.security = d.security
  }

  const prevStep = () => {
    if (activeStep.value > 0) activeStep.value--
  }

  const nextStep = async () => {
    if (activeStep.value === 0) {
      if (!provider.value) {
        ElMessage.warning('请选择邮箱类型')
        return
      }
      activeStep.value = 1
      return
    }
    if (activeStep.value === 1) {
      await configFormRef.value?.validate()
      if (!form.displayName && form.email) {
        form.displayName = form.email.split('@')[0] || ''
      }
      testState.status = ''
      testState.message = ''
      activeStep.value = 2
      return
    }
  }

  const testConnection = async () => {
    if (testState.loading) return
    testState.loading = true
    testState.status = ''
    testState.message = ''
    try {
      if (mode.value === 'edit' && editingId.value && !form.secret) {
        const res = await testSenderEmailAccountById({ id: editingId.value })
        if (res.code === 0) {
          testState.status = 'success'
          testState.message = res.msg || '连接成功，测试邮件已发送'
          return
        }
        testState.status = 'error'
        testState.message = res.msg || '测试失败'
        return
      }
      const res = await testSenderEmailAccount({
        email: form.email,
        displayName: form.displayName,
        smtpHost: form.smtpHost,
        smtpPort: form.smtpPort,
        security: form.security,
        isLoginAuth: form.isLoginAuth,
        secret: form.secret,
      })
      if (res.code === 0) {
        testState.status = 'success'
        testState.message = res.msg || '连接成功，测试邮件已发送'
      } else {
        testState.status = 'error'
        testState.message = res.msg || '测试失败'
      }
    } catch (e) {
      testState.status = 'error'
      testState.message = e?.message || '测试失败'
    } finally {
      testState.loading = false
    }
  }

  const saveAccount = async () => {
    if (saving.value) return
    if (testState.status !== 'success') {
      ElMessage.warning('请先测试连接成功后再保存')
      return
    }
    saving.value = true
    try {
      if (mode.value === 'create') {
        const res = await createSenderEmailAccount({
          email: form.email,
          displayName: form.displayName,
          provider: provider.value || 'custom',
          smtpHost: form.smtpHost,
          smtpPort: form.smtpPort,
          security: form.security,
          isLoginAuth: form.isLoginAuth,
          secret: form.secret,
          dailyLimit: form.dailyLimit,
          intervalPolicy: form.intervalPolicy,
        })
        if (res.code === 0) {
          ElMessage.success('创建成功')
          drawerVisible.value = false
          await fetchList()
        }
        return
      }
      const res = await updateSenderEmailAccount({
        id: editingId.value,
        email: form.email,
        displayName: form.displayName,
        provider: provider.value || 'custom',
        smtpHost: form.smtpHost,
        smtpPort: form.smtpPort,
        security: form.security,
        isLoginAuth: form.isLoginAuth,
        secret: form.secret,
        dailyLimit: form.dailyLimit,
        intervalPolicy: form.intervalPolicy,
      })
      if (res.code === 0) {
        ElMessage.success('更新成功')
        drawerVisible.value = false
        await fetchList()
      }
    } finally {
      saving.value = false
    }
  }

  const setDefault = async (row) => {
    const res = await setDefaultSenderEmailAccount({ id: row.id })
    if (res.code === 0) {
      ElMessage.success('设置成功')
      fetchList()
    }
  }

  const testById = async (row) => {
    row.__testing = true
    try {
      const res = await testSenderEmailAccountById({ id: row.id })
      if (res.code === 0) {
        ElMessage.success(res.msg || '连接成功，测试邮件已发送')
        fetchList()
      }
    } finally {
      row.__testing = false
    }
  }

  const removeRow = async (row) => {
    await ElMessageBox.confirm('确定要删除该邮箱账号吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await deleteSenderEmailAccount({ id: row.id })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      fetchList()
    }
  }

  const quotaDialogVisible = ref(false)
  const quotaSaving = ref(false)
  const quotaForm = reactive({
    id: 0,
    dailyLimit: 50,
    intervalPolicy: '1m',
  })

  const openQuota = (row) => {
    quotaForm.id = row.id
    quotaForm.dailyLimit = row.dailyLimit
    quotaForm.intervalPolicy = row.intervalPolicy || '1m'
    quotaDialogVisible.value = true
  }

  const saveQuota = async () => {
    if (quotaSaving.value) return
    quotaSaving.value = true
    try {
      const res = await updateSenderEmailAccountQuota({
        id: quotaForm.id,
        dailyLimit: quotaForm.dailyLimit,
        intervalPolicy: quotaForm.intervalPolicy,
      })
      if (res.code === 0) {
        ElMessage.success('更新成功')
        quotaDialogVisible.value = false
        fetchList()
      }
    } finally {
      quotaSaving.value = false
    }
  }

  fetchList()
</script>
