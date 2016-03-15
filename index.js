'use strict'
const exec = require('child_process').exec
const crypto = require('crypto')
const fs = require('fs')

const _ = require('lodash')
const boom = require('boom')
const hapi = require('hapi')
const client = require('request')

const server = new hapi.Server()

server.connection({
  port: process.env.PORT || 5000
})

const validName = /^[a-zA-Z0-9-_/]+$/
const deploysh = _.template(fs.readFileSync('deploy.sh', 'utf8'))
const defaultTag = process.env.DEFAULT_TAG || 'latest'
const defaultToken = process.env.DEFAULT_TOKEN || crypto.randomBytes(12).toString('hex')
const defaultParams = _.template(process.env.DEFAULT_PARAMS || '')

console.log('default tag:', defaultTag)
console.log('default token:', defaultToken)
console.log('default params:', defaultParams({
  repo_name: '<repo_name>',
  name: '<name>'
}))

server.route({
  method: 'POST',
  path: '/{token*}',
  handler: (request, reply) => {
    if (request.params.token !== defaultToken) {
      return reply(boom.unauthorized('invalid token'))
    }

    const payload = request.payload || {}
    const repo = payload.repository || {}
    if (!payload.callback_url ||
        !validName.test(repo.repo_name || '') ||
        !validName.test(repo.name || '')) {
      return reply(boom.badRequest('invalid input'))
    }

    function sendCallback (success, description) {
      client.post({
        url: payload.callback_url,
        json: true,
        body: {
          state: success ? 'success' : 'failure',
          context: 'Webhook deploy server',
          description: (description || '').substr(0, 255)
        }
      }, (err, res, data) => {
        if (err || res.statusCode !== 200) {
          return reply(boom.badRequest('invalid callback'))
        }
        reply({ok: true})
      })
    }

    if (_.get(payload, 'push_data.tag') !== defaultTag) {
      return sendCallback(true, 'skipped tag')
    }

    const info = {
      repo_name: repo.repo_name,
      name: repo.name,
      tag: defaultTag
    }
    const script = deploysh(_.assign({
      params: defaultParams(info)
    }, info))
    exec(`bash -c '${script}' 2>&1`, (error, stdout, stderr) => {
      if (error) return sendCallback(false, 'script error: ' + error)
      sendCallback(true, 'successfully deployed image:\n' + stdout.toString())
    })
  }
})

server.register({
  register: require('good'),
  options: {
    reporters: [{
      reporter: require('good-console'),
      events: { error: '*', log: '*', response: '*' }
    }]
  }
}, (err) => {
  if (err) return console.log(err)
  server.start(() => console.log('server started:', server.info.uri))
})
