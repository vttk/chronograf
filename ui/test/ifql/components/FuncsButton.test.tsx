import React from 'react'
import {shallow} from 'enzyme'
import {FuncsButton} from 'src/ifql/components/FuncsButton'
import DropdownInput from 'src/shared/components/DropdownInput'

const setup = (override = {}) => {
  const props = {
    funcs: ['f1', 'f2'],
    ...override,
  }

  const wrapper = shallow(<FuncsButton {...props} />)

  return {
    wrapper,
  }
}

describe('IFQL.Components.FuncsButton', () => {
  describe('rendering', () => {
    it('renders', () => {
      const {wrapper} = setup()

      expect(wrapper.exists()).toBe(true)
    })

    describe('the function list', () => {
      it('does not render the list of funcs by default', () => {
        const {wrapper} = setup()

        const list = wrapper.find({'data-test': 'func-li'})

        expect(list.exists()).toBe(false)
      })
    })
  })

  describe('user interraction', () => {
    describe('clicking the add function button', () => {
      it('displays the list of functions', () => {
        const {wrapper} = setup()

        const dropdownButton = wrapper.find('button')
        dropdownButton.simulate('click')

        const list = wrapper.find('.func')

        expect(list.length).toBe(2)
        expect(list.first().text()).toBe('f1')
        expect(list.last().text()).toBe('f2')
      })
    })

    describe('filtering the list', () => {
      it('displays the filtered funcs', () => {
        const {wrapper} = setup()

        const dropdownButton = wrapper.find('button')
        dropdownButton.simulate('click')

        let list = wrapper.find('.func')

        expect(list.length).toBe(2)
        expect(list.first().text()).toBe('f1')
        expect(list.last().text()).toBe('f2')

        const input = wrapper
          .find(DropdownInput)
          .dive()
          .find('input')

        input.simulate('change', {target: {value: '2'}})
        wrapper.update()

        list = wrapper.find('.func')

        expect(list.length).toBe(1)
        expect(list.first().text()).toBe('f2')
      })
    })

    describe('exiting the list', () => {
      it('closes when ESC is pressed', () => {
        const {wrapper} = setup()

        const dropdownButton = wrapper.find('button')
        dropdownButton.simulate('click')
        let list = wrapper.find('.func')

        expect(list.exists()).toBe(true)

        const input = wrapper
          .find(DropdownInput)
          .dive()
          .find('input')

        input.simulate('keyDown', {key: 'Escape'})
        wrapper.update()

        list = wrapper.find('.func')

        expect(list.exists()).toBe(false)
      })
    })
  })
})