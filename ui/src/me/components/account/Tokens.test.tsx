// Libraries
import React from 'react'
import {mount} from 'enzyme'

// Components
import {Tokens} from 'src/me/components/account/Tokens'
import TokenRow from 'src/me/components/account/TokenRow'

jest.mock('src/authorizations/apis')

const setup = (override?) => {
  const props = {
    authorizationsLink: 'api/v2/authorizations',
    ...override,
  }

  const tokensWrapper = mount(<Tokens {...props} />)

  return {tokensWrapper}
}

describe('Account', () => {
  let wrapper
  beforeEach(done => {
    const {tokensWrapper} = setup()
    wrapper = tokensWrapper
    process.nextTick(() => {
      wrapper.update()
      done()
    })
  })

  describe('rendering', () => {
    it('renders!', () => {
      expect(wrapper.exists()).toBe(true)
      expect(wrapper).toMatchSnapshot()
    })

    it('displays the list of tokens', () => {
      const rows = wrapper.find(TokenRow)
      expect(rows.length).toBe(2)
    })
  })

  // describe('user interaction', () => {
  //   describe('clicking the token description', () => {
  //     it('opens the ViewTokenModal', () => {
  //       const {wrapper} = setup()
  //     })
  //   })
  // })
})
