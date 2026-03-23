import service from '@/utils/request'

export const getContactList = (params) => {
  return service({
    url: '/contact/list',
    method: 'get',
    params
  })
}

export const createContact = (data) => {
  return service({
    url: '/contact/create',
    method: 'post',
    data
  })
}

