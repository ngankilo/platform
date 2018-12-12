// Reducer
import dataLoadersReducer, {
// DataLoadersState,
  INITIAL_STATE,
} from 'src/onboarding/reducers/dataLoaders'

// Actions
import {
  setDataLoadersType,
  addTelegrafPlugin,
  addConfigValue,
  removeConfigValue,
  removeTelegrafPlugin,
  removePluginBundle,
  setActiveTelegrafPlugin,
  setTelegrafConfigID,
  updateTelegrafPluginConfig,
  addPluginBundle,
  addTelegrafPlugins,
  removeBundlePlugins,
} from 'src/onboarding/actions/dataLoaders'

// Types
import {
  TelegrafPluginInputCpu,
  TelegrafPluginInputDisk,
  TelegrafPluginInputRedis,
  TelegrafPluginInputKernel,
  TelegrafPluginInputDiskio,
  TelegrafPluginInputMem,
  TelegrafPluginInputSwap,
  TelegrafPluginInputSystem,
  TelegrafPluginInputProcesses,
} from 'src/api'
import {
  DataLoaderType,
  ConfigurationState,
  TelegrafPlugin,
  BundleName,
} from 'src/types/v2/dataLoaders'
import {redisPlugin, telegrafPlugin} from 'mocks/dummyData'

describe('dataLoader reducer', () => {
  it('can set a type', () => {
    const actual = dataLoadersReducer(
      INITIAL_STATE,
      setDataLoadersType(DataLoaderType.Streaming)
    )
    const expected = {
      ...INITIAL_STATE,
      telegrafPlugins: [],
      type: DataLoaderType.Streaming,
    }

    expect(actual).toEqual(expected)
  })

  it('can add a telegraf plugin', () => {
    const actual = dataLoadersReducer(
      INITIAL_STATE,
      addTelegrafPlugin({
        name: TelegrafPluginInputCpu.NameEnum.Cpu,
        configured: ConfigurationState.Unconfigured,
        active: true,
      })
    )
    const expected = {
      ...INITIAL_STATE,
      telegrafPlugins: [
        {
          name: TelegrafPluginInputCpu.NameEnum.Cpu,
          configured: ConfigurationState.Unconfigured,
          active: true,
        },
      ],
      type: DataLoaderType.Empty,
    }

    expect(actual).toEqual(expected)
  })

  it('can set the active telegraf plugin', () => {
    const actual = dataLoadersReducer(
      {
        ...INITIAL_STATE,
        type: DataLoaderType.Streaming,
        telegrafPlugins: [
          {
            name: TelegrafPluginInputCpu.NameEnum.Cpu,
            configured: ConfigurationState.Unconfigured,
            active: true,
          },
          {
            name: TelegrafPluginInputDisk.NameEnum.Disk,
            configured: ConfigurationState.Unconfigured,
            active: false,
          },
        ],
      },
      setActiveTelegrafPlugin(TelegrafPluginInputDisk.NameEnum.Disk)
    )

    const expected = {
      ...INITIAL_STATE,
      type: DataLoaderType.Streaming,
      telegrafPlugins: [
        {
          name: TelegrafPluginInputCpu.NameEnum.Cpu,
          configured: ConfigurationState.Unconfigured,
          active: false,
        },
        {
          name: TelegrafPluginInputDisk.NameEnum.Disk,
          configured: ConfigurationState.Unconfigured,
          active: true,
        },
      ],
    }

    expect(actual).toEqual(expected)
  })

  it('can remove a telegraf plugin', () => {
    const actual = dataLoadersReducer(
      {
        ...INITIAL_STATE,
        type: DataLoaderType.Streaming,
        telegrafPlugins: [
          {
            name: TelegrafPluginInputCpu.NameEnum.Cpu,
            configured: ConfigurationState.Unconfigured,
            active: true,
          },
        ],
      },
      removeTelegrafPlugin(TelegrafPluginInputCpu.NameEnum.Cpu)
    )
    const expected = {
      ...INITIAL_STATE,
      telegrafPlugins: [],
      type: DataLoaderType.Streaming,
    }

    expect(actual).toEqual(expected)
  })

  it('can set a telegraf config id', () => {
    const id = '285973845720345ajfajfkl;'
    const actual = dataLoadersReducer(INITIAL_STATE, setTelegrafConfigID(id))

    const expected = {...INITIAL_STATE, telegrafConfigID: id}

    expect(actual).toEqual(expected)
  })

  it('can update a plugin config field', () => {
    const plugin = {
      ...redisPlugin,
      config: {servers: [], password: ''},
    }
    const tp: TelegrafPlugin = {
      name: TelegrafPluginInputRedis.NameEnum.Redis,
      configured: ConfigurationState.Unconfigured,
      active: true,
      plugin,
    }
    const actual = dataLoadersReducer(
      {...INITIAL_STATE, telegrafPlugins: [tp]},
      updateTelegrafPluginConfig(
        TelegrafPluginInputRedis.NameEnum.Redis,
        'password',
        'pa$$w0rd'
      )
    )

    const expected = {
      ...INITIAL_STATE,
      telegrafPlugins: [
        {
          ...tp,
          plugin: {
            ...plugin,
            config: {servers: [], password: 'pa$$w0rd'},
          },
        },
      ],
    }

    expect(actual).toEqual(expected)
  })

  it('can add a plugin config value', () => {
    const plugin = {
      ...redisPlugin,
      config: {servers: ['first'], password: ''},
    }
    const tp: TelegrafPlugin = {
      name: TelegrafPluginInputRedis.NameEnum.Redis,
      configured: ConfigurationState.Unconfigured,
      active: true,
      plugin,
    }
    const actual = dataLoadersReducer(
      {...INITIAL_STATE, telegrafPlugins: [tp]},
      addConfigValue(
        TelegrafPluginInputRedis.NameEnum.Redis,
        'servers',
        'second'
      )
    )

    const expected = {
      ...INITIAL_STATE,
      telegrafPlugins: [
        {
          ...tp,
          plugin: {
            ...plugin,
            config: {servers: ['first', 'second'], password: ''},
          },
        },
      ],
    }

    expect(actual).toEqual(expected)
  })

  it('can remove a plugin config value', () => {
    const plugin = {
      ...redisPlugin,
      config: {servers: ['first', 'second'], password: ''},
    }
    const tp: TelegrafPlugin = {
      name: TelegrafPluginInputRedis.NameEnum.Redis,
      configured: ConfigurationState.Unconfigured,
      active: true,
      plugin,
    }
    const actual = dataLoadersReducer(
      {...INITIAL_STATE, telegrafPlugins: [tp]},
      removeConfigValue(
        TelegrafPluginInputRedis.NameEnum.Redis,
        'servers',
        'first'
      )
    )

    const expected = {
      ...INITIAL_STATE,
      telegrafPlugins: [
        {
          ...tp,
          plugin: {
            ...plugin,
            config: {servers: ['second'], password: ''},
          },
        },
      ],
    }
    expect(actual).toEqual(expected)
  })

  it('can add a plugin bundle', () => {
    const actual = dataLoadersReducer(
      {...INITIAL_STATE, pluginBundles: [BundleName.Redis]},
      addPluginBundle(BundleName.System)
    )

    const expected = {
      ...INITIAL_STATE,
      pluginBundles: [BundleName.Redis, BundleName.System],
    }
    expect(actual).toEqual(expected)
  })

  it('can remove a plugin bundle', () => {
    const actual = dataLoadersReducer(
      {...INITIAL_STATE, pluginBundles: [BundleName.Redis, BundleName.System]},
      removePluginBundle(BundleName.Redis)
    )

    const expected = {
      ...INITIAL_STATE,
      pluginBundles: [BundleName.System],
    }
    expect(actual).toEqual(expected)
  })

  it('can add telegraf Plugins', () => {
    const cpuPlugin = {...telegrafPlugin}
    const diskPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputDisk.NameEnum.Disk,
    }
    const redisPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputRedis.NameEnum.Redis,
    }

    const actual = dataLoadersReducer(
      {...INITIAL_STATE, telegrafPlugins: [redisPlugin, diskPlugin]},
      addTelegrafPlugins([cpuPlugin, diskPlugin])
    )

    const expected = {
      ...INITIAL_STATE,
      telegrafPlugins: [redisPlugin, diskPlugin, cpuPlugin],
    }

    expect(actual).toEqual(expected)
  })

  it('can remove telegraf Plugins', () => {
    const cpuPlugin = {...telegrafPlugin}
    const diskPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputDisk.NameEnum.Disk,
    }
    const diskioPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputDiskio.NameEnum.Diskio,
    }
    const kernelPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputKernel.NameEnum.Kernel,
    }
    const memPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputMem.NameEnum.Mem,
    }
    const processesPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputProcesses.NameEnum.Processes,
    }
    const swapPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputSwap.NameEnum.Swap,
    }
    const systemPlugin = {
      ...telegrafPlugin,
      name: TelegrafPluginInputSystem.NameEnum.System,
    }

    const actual = dataLoadersReducer(
      {
        ...INITIAL_STATE,
        pluginBundles: [BundleName.Disk, BundleName.System],
        telegrafPlugins: [
          cpuPlugin,
          diskPlugin,
          diskioPlugin,
          kernelPlugin,
          memPlugin,
          processesPlugin,
          swapPlugin,
          systemPlugin,
        ],
      },
      removeBundlePlugins(BundleName.System)
    )

    const expected = {
      ...INITIAL_STATE,
      pluginBundles: [BundleName.Disk, BundleName.System],
      telegrafPlugins: [diskPlugin, diskioPlugin],
    }

    expect(actual).toEqual(expected)
  })
})
