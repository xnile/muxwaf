import storage from 'store'

const site = {
  state: {
    id: 0,
    domain: ''
  },

  mutations: {
    SET_ID: (state, id) => {
      state.id = id
    },
    SET_DOMAIN: (state, domain) => {
      state.domain = domain
    }
  },

  actions: {}
}

export default site
