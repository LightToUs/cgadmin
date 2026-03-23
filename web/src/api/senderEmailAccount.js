import service from '@/utils/request'

export const getSenderEmailAccountList = () => {
  return service({
    url: '/senderEmailAccount/list',
    method: 'get'
  })
}

export const createSenderEmailAccount = (data) => {
  return service({
    url: '/senderEmailAccount/create',
    method: 'post',
    data
  })
}

export const updateSenderEmailAccount = (data) => {
  return service({
    url: '/senderEmailAccount/update',
    method: 'put',
    data
  })
}

export const updateSenderEmailAccountQuota = (data) => {
  return service({
    url: '/senderEmailAccount/updateQuota',
    method: 'put',
    data
  })
}

export const deleteSenderEmailAccount = (data) => {
  return service({
    url: '/senderEmailAccount/delete',
    method: 'delete',
    data
  })
}

export const setDefaultSenderEmailAccount = (data) => {
  return service({
    url: '/senderEmailAccount/setDefault',
    method: 'post',
    data
  })
}

export const testSenderEmailAccount = (data) => {
  return service({
    url: '/senderEmailAccount/test',
    method: 'post',
    data
  })
}

export const testSenderEmailAccountById = (data) => {
  return service({
    url: '/senderEmailAccount/testById',
    method: 'post',
    data
  })
}

