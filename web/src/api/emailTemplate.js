import service from '@/utils/request'

export const getEmailTemplateList = (params) => {
  return service({
    url: '/emailTemplate/list',
    method: 'get',
    params
  })
}

export const getEmailTemplateDetail = (params) => {
  return service({
    url: '/emailTemplate/detail',
    method: 'get',
    params
  })
}

export const createEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/create',
    method: 'post',
    data
  })
}

export const updateEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/update',
    method: 'put',
    data
  })
}

export const deleteEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/delete',
    method: 'delete',
    data
  })
}

export const deleteEmailTemplateByIds = (data) => {
  return service({
    url: '/emailTemplate/deleteByIds',
    method: 'delete',
    data
  })
}

export const copyEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/copy',
    method: 'post',
    data
  })
}

export const batchStatusEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/batchStatus',
    method: 'post',
    data
  })
}

export const moveEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/move',
    method: 'post',
    data
  })
}

export const previewEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/preview',
    method: 'post',
    data
  })
}

export const testSendEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/testSend',
    method: 'post',
    data
  })
}

export const exportEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/export',
    method: 'post',
    data
  })
}

export const importEmailTemplate = (data) => {
  return service({
    url: '/emailTemplate/import',
    method: 'post',
    headers: { 'Content-Type': 'multipart/form-data' },
    data
  })
}
