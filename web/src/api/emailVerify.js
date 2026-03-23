import service from '@/utils/request'

export const startEmailVerify = (data) => {
  return service({
    url: '/emailVerify/start',
    method: 'post',
    data
  })
}

export const getEmailVerifyJob = (params) => {
  return service({
    url: '/emailVerify/job',
    method: 'get',
    params
  })
}

export const getEmailVerifyHistory = (params) => {
  return service({
    url: '/emailVerify/history',
    method: 'get',
    params
  })
}

