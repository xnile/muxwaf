import request from '@/utils/request'

export function GetAll(params) {
  return request({
    url: '/api/certificates/all',
    method: 'get',
    params: params
  })
}

export function AddCert(data) {
  return request({
    url: '/api/certificates',
    method: 'post',
    data: data
  })
}

export function UpdateCert(id, data) {
  return request({
    url: `/api/certificates/${id}`,
    method: 'put',
    data: data
  })
}

export function ListCert(params) {
  return request({
    url: '/api/certificates',
    method: 'get',
    params: params
  })
}

export function DelCert(id) {
  return request({
    url: `/api/certificates/${id}`,
    method: 'delete'
  })
}

export function GetDomainCert(domain) {
  return request({
    url: `/api/certificates`,
    method: 'get',
    params: {
      domain: domain
    }
  })
}

// export function GetCertName(id) {
//   return request({
//     url: `/api/certificates/${id}/name`,
//     method: 'get'
//   })
// }
