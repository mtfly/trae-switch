<script>
  import { onMount } from 'svelte'
  import Button from './components/ui/Button.svelte'
  import Card from './components/ui/Card.svelte'
  import Switch from './components/ui/Switch.svelte'

  let status = {
    runningAsAdmin: false,
    hostsSet: false,
    certInstalled: false,
    proxyRunning: false,
    proxyPort: 443,
    portAvailable: true,
    portProcess: '',
    activeProvider: null,
    activeTargetURL: ''
  }

  let providers = []
  let activeProviderIndex = 0
  let loading = false
  let error = ''
  let success = ''
  let theme = 'dark'
  let trayMode = false

  let showProviderModal = false
  let editingProviderIndex = -1
  let providerForm = {
    name: '',
    openaiBase: '',
    models: []
  }
  let newModelInput = ''

  const themes = [
    { value: 'dark', label: '黑' },
    { value: 'light', label: '白' },
    { value: 'gray', label: '灰' }
  ]

  function applyTheme(nextTheme) {
    const root = document.documentElement
    root.classList.remove('dark', 'theme-gray')

    if (nextTheme === 'dark') {
      root.classList.add('dark')
    }

    if (nextTheme === 'gray') {
      root.classList.add('theme-gray')
    }

    theme = nextTheme
    localStorage.setItem('theme', nextTheme)
  }

  onMount(async () => {
    const savedTheme = localStorage.getItem('theme')
    applyTheme(themes.some((item) => item.value === savedTheme) ? savedTheme : 'dark')

    await refreshStatus()
    await loadProviders()
    await loadTrayMode()
  })

  async function loadTrayMode() {
    try {
      trayMode = await window.go.main.App.GetTrayMode()
    } catch (e) {
      console.error('Failed to load tray mode:', e)
    }
  }

  async function toggleTrayMode() {
    try {
      await window.go.main.App.SetTrayMode(trayMode)
      if (trayMode) {
        showSuccess('常驻任务栏已开启')
      } else {
        showSuccess('常驻任务栏已关闭')
      }
    } catch (e) {
      showError(e.message || String(e))
    }
  }

  function toggleTheme() {
    const currentIndex = themes.findIndex((item) => item.value === theme)
    const nextTheme = themes[(currentIndex + 1) % themes.length].value
    applyTheme(nextTheme)
  }

  async function refreshStatus() {
    try {
      const result = await window.go.main.App.GetStatus()
      status = result
    } catch (e) {
      console.error('Failed to get status:', e)
    }
  }

  async function loadProviders() {
    try {
      providers = await window.go.main.App.GetProviders()
      activeProviderIndex = await window.go.main.App.GetActiveProviderIndex()
    } catch (e) {
      console.error('Failed to load providers:', e)
    }
  }

  function showError(msg) {
    success = ''
    error = msg
    setTimeout(() => {
      error = ''
    }, 5000)
  }

  function showSuccess(msg) {
    error = ''
    success = msg
    setTimeout(() => {
      success = ''
    }, 3000)
  }

  async function setHosts() {
    loading = true
    try {
      await window.go.main.App.SetHosts()
      await refreshStatus()
      showSuccess('Hosts 已设置')
    } catch (e) {
      showError(e.message || String(e))
    }
    loading = false
  }

  async function restoreHosts() {
    loading = true
    try {
      await window.go.main.App.RestoreHosts()
      await refreshStatus()
      showSuccess('Hosts 已恢复')
    } catch (e) {
      showError(e.message || String(e))
    }
    loading = false
  }

  async function installCert() {
    loading = true
    try {
      await window.go.main.App.InstallCertificate()
      await refreshStatus()
      showSuccess('CA 证书已安装')
    } catch (e) {
      showError(e.message || String(e))
    }
    loading = false
  }

  async function uninstallCert() {
    loading = true
    try {
      await window.go.main.App.UninstallCertificate()
      await refreshStatus()
      showSuccess('CA 证书已卸载')
    } catch (e) {
      showError(e.message || String(e))
    }
    loading = false
  }

  async function startProxy() {
    loading = true
    try {
      await window.go.main.App.StartProxy()
      await refreshStatus()
      showSuccess('代理已启动')
    } catch (e) {
      showError(e.message || String(e))
    }
    loading = false
  }

  async function stopProxy() {
    loading = true
    try {
      await window.go.main.App.StopProxy()
      await refreshStatus()
      showSuccess('代理已停止')
    } catch (e) {
      showError(e.message || String(e))
    }
    loading = false
  }

  function getStartReminder() {
    if (!status.hostsSet && !status.certInstalled) {
      return '请先完成 Hosts 配置并安装 CA 证书后再启动'
    }

    if (!status.hostsSet) {
      return '请先完成 Hosts 配置后再启动'
    }

    if (!status.certInstalled) {
      return '请先安装 CA 证书后再启动'
    }

    return ''
  }

  async function handlePrimaryAction() {
    if (status.proxyRunning) {
      await stopProxy()
      return
    }

    const reminder = getStartReminder()
    if (reminder) {
      showError(reminder)
      return
    }

    await startProxy()
  }

  async function selectProvider(index) {
    try {
      await window.go.main.App.SetActiveProvider(index)
      activeProviderIndex = index
      await refreshStatus()
      showSuccess('已切换服务商')
    } catch (e) {
      showError(e.message || String(e))
    }
  }

  function openAddProvider() {
    editingProviderIndex = -1
    providerForm = { name: '', openaiBase: '', models: [] }
    newModelInput = ''
    showProviderModal = true
  }

  function closeProviderModal() {
    showProviderModal = false
  }

  function openEditProvider(index) {
    const provider = providers[index]
    editingProviderIndex = index
    providerForm = {
      name: provider.name,
      openaiBase: provider.openai_base || '',
      models: [...(provider.models || [])]
    }
    newModelInput = ''
    showProviderModal = true
  }

  function addModel() {
    const model = newModelInput.trim()
    if (model && !providerForm.models.includes(model)) {
      providerForm.models = [...providerForm.models, model]
      newModelInput = ''
    }
  }

  function removeModel(index) {
    providerForm.models = providerForm.models.filter((_, i) => i !== index)
  }

  async function saveProvider() {
    if (!providerForm.name || !providerForm.openaiBase) {
      showError('请填写名称和 API 地址')
      return
    }

    try {
      if (editingProviderIndex >= 0) {
        await window.go.main.App.UpdateProvider(
          editingProviderIndex,
          providerForm.name,
          providerForm.openaiBase,
          providerForm.models
        )
        showSuccess('服务商已更新')
      } else {
        await window.go.main.App.AddProvider(
          providerForm.name,
          providerForm.openaiBase,
          providerForm.models
        )
        showSuccess('服务商已添加')
      }

      showProviderModal = false
      await loadProviders()
    } catch (e) {
      showError(e.message || String(e))
    }
  }

  async function deleteProvider(index) {
    if (!confirm('确定要删除该服务商吗？')) return

    try {
      await window.go.main.App.DeleteProvider(index)
      showSuccess('服务商已删除')
      await loadProviders()
      await refreshStatus()
    } catch (e) {
      showError(e.message || String(e))
    }
  }

  $: adminWarning = !status.runningAsAdmin
  $: portWarning = !status.portAvailable && !status.proxyRunning
  $: primaryButtonDisabled = loading || (!status.proxyRunning && (adminWarning || portWarning))
  $: primaryButtonLabel = status.proxyRunning ? '停止' : '启动'
  $: primaryButtonVariant = status.proxyRunning ? 'destructive' : 'default'
  $: displayTargetURL = status.activeTargetURL || '未设置'
  $: activeProviderName = status.activeProvider ? status.activeProvider.name : '未选择'
  $: themeButtonTitle =
    theme === 'dark'
      ? '当前黑色主题，点击切换到白色主题'
      : theme === 'light'
        ? '当前白色主题，点击切换到灰色主题'
        : '当前灰色主题，点击切换到黑色主题'
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="mx-auto max-w-md space-y-4 p-4">
    <div class="grid grid-cols-4 gap-3 py-2">
      <Button
        variant={primaryButtonVariant}
        size="lg"
        on:click={handlePrimaryAction}
        disabled={primaryButtonDisabled}
        className="col-span-3 w-full"
      >
        {primaryButtonLabel}
      </Button>

      <button
        type="button"
        on:click={toggleTheme}
        class="col-span-1 flex h-11 items-center justify-center gap-1 rounded-lg border bg-card px-2 text-card-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
        title={themeButtonTitle}
        aria-label={themeButtonTitle}
      >
        {#each themes as item}
          <span
            class={`rounded px-1.5 py-0.5 text-xs font-medium transition-colors ${
              theme === item.value
                ? 'bg-foreground text-background'
                : 'text-muted-foreground'
            }`}
          >
            {item.label}
          </span>
        {/each}
      </button>
    </div>

    {#if error}
      <div class="rounded-lg border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
        {error}
      </div>
    {/if}

    {#if success}
      <div class="rounded-lg border border-success/20 bg-success/10 p-3 text-sm text-success">
        {success}
      </div>
    {/if}

    {#if adminWarning}
      <div class="rounded-lg border border-yellow-500/20 bg-yellow-500/10 p-3 text-sm text-yellow-600 dark:text-yellow-500">
        需要以管理员权限运行应用后才能监听 443 端口。
      </div>
    {/if}

    {#if portWarning}
      <div class="rounded-lg border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
        443 端口已被占用，请先关闭占用该端口的程序。
        {#if status.portProcess}
          <span class="mt-1 block text-xs opacity-80">{status.portProcess}</span>
        {/if}
      </div>
    {/if}

    <Card className="p-4">
      <div class="flex items-center gap-3">
        {#if status.proxyRunning}
          <span class="h-3 w-3 rounded-full bg-success animate-pulse-slow"></span>
        {:else}
          <span class="h-3 w-3 rounded-full bg-muted-foreground"></span>
        {/if}
        <span class="font-medium">{status.proxyRunning ? '代理运行中' : '代理已停止'}</span>
      </div>

      <div class="mt-3 space-y-1 text-xs text-muted-foreground">
        <div>当前服务商：{activeProviderName}</div>
        <div class="break-all">当前目标：{displayTargetURL}</div>
      </div>
    </Card>

    <Card className="p-4">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="font-semibold">服务商</h2>
        <Button variant="ghost" size="sm" on:click={openAddProvider}>+ 添加</Button>
      </div>

      {#if providers.length === 0}
        <div class="py-4 text-center text-sm text-muted-foreground">
          暂无服务商，请先添加。
        </div>
      {:else}
        <div class="space-y-2">
          {#each providers as provider, i}
            <div
              class="flex items-center justify-between rounded-lg bg-muted/50 p-3 transition-colors hover:bg-muted {i === activeProviderIndex ? 'bg-primary/10' : ''}"
              on:click={() => selectProvider(i)}
              on:keydown={(event) => (event.key === 'Enter' || event.key === ' ') && selectProvider(i)}
              role="button"
              tabindex="0"
            >
              <div class="flex min-w-0 items-center gap-3">
                {#if i === activeProviderIndex}
                  <span class="h-2 w-2 rounded-full bg-primary"></span>
                {:else}
                  <span class="h-2 w-2 rounded-full bg-muted-foreground"></span>
                {/if}

                <div class="min-w-0">
                  <div class="text-sm font-medium">{provider.name}</div>
                  <div class="max-w-[180px] break-all text-xs text-muted-foreground">
                    {provider.openai_base}
                  </div>
                </div>
              </div>

              <div
                class="flex gap-1"
                on:click|stopPropagation
                on:keydown|stopPropagation
                role="presentation"
              >
                <Button variant="ghost" size="sm" on:click={() => openEditProvider(i)}>编辑</Button>
                <Button variant="ghost" size="sm" on:click={() => deleteProvider(i)}>删除</Button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </Card>

    <Card className="p-4">
      <h2 class="mb-4 font-semibold">系统配置</h2>

      <div class="mb-3 flex items-center justify-between rounded-lg bg-muted/50 p-3">
        <div class="flex items-center gap-3">
          {#if trayMode}
            <span class="h-2 w-2 rounded-full bg-success"></span>
          {:else}
            <span class="h-2 w-2 rounded-full bg-muted-foreground"></span>
          {/if}
          <div>
            <div class="text-sm font-medium">常驻任务栏</div>
            <div class="text-xs text-muted-foreground">
              {trayMode ? '已开启' : '已关闭'}
            </div>
          </div>
        </div>
        <Switch checked={trayMode} on:change={() => { trayMode = !trayMode; toggleTrayMode() }} />
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div class="rounded-lg bg-muted/50 p-3">
          <div class="flex items-center gap-3">
            {#if status.hostsSet}
              <span class="h-2 w-2 rounded-full bg-success"></span>
            {:else}
              <span class="h-2 w-2 rounded-full bg-muted-foreground"></span>
            {/if}

            <div>
              <div class="text-sm font-medium">Hosts 配置</div>
              <div class="text-xs text-muted-foreground">
                {status.hostsSet ? '已设置' : '未设置'}
              </div>
            </div>
          </div>

          <div class="mt-3 flex flex-wrap gap-2">
            <Button variant="ghost" size="sm" on:click={setHosts} disabled={loading || status.hostsSet}>
              设置
            </Button>
            <Button variant="ghost" size="sm" on:click={restoreHosts} disabled={loading || !status.hostsSet}>
              恢复
            </Button>
          </div>
        </div>

        <div class="rounded-lg bg-muted/50 p-3">
          <div class="flex items-center gap-3">
            {#if status.certInstalled}
              <span class="h-2 w-2 rounded-full bg-success"></span>
            {:else}
              <span class="h-2 w-2 rounded-full bg-muted-foreground"></span>
            {/if}

            <div>
              <div class="text-sm font-medium">CA 证书</div>
              <div class="text-xs text-muted-foreground">
                {status.certInstalled ? '已安装' : '未安装'}
              </div>
            </div>
          </div>

          <div class="mt-3 flex flex-wrap gap-2">
            <Button variant="ghost" size="sm" on:click={installCert} disabled={loading || status.certInstalled}>
              安装
            </Button>
            <Button variant="ghost" size="sm" on:click={uninstallCert} disabled={loading || !status.certInstalled}>
              卸载
            </Button>
          </div>
        </div>
      </div>
    </Card>

    <Card className="p-4">
      <h2 class="mb-3 font-semibold">使用说明</h2>
      <ol class="list-decimal space-y-2 pl-5 text-sm leading-6 text-muted-foreground">
        <li>添加服务商配置（API 地址和模型列表）并点击「启动」</li>
        <li>在 Trae IDE 添加自定义模型 服务商选择OpenAI 服务商</li>
        <li>模型手动输入你想要使用的模型并且输入对应 API Key</li>
        <li>关闭auto mode 并且选择刚添加的模型</li>
      </ol>
    </Card>
  </div>
</div>

{#if showProviderModal}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    on:click|self={closeProviderModal}
    on:keydown={(event) => event.key === 'Escape' && closeProviderModal()}
    role="presentation"
    tabindex="-1"
  >
    <div
      class="mx-4 w-full max-w-md rounded-lg bg-card p-6"
      role="dialog"
      aria-modal="true"
    >
      <h3 class="mb-4 text-lg font-semibold">
        {editingProviderIndex >= 0 ? '编辑服务商' : '添加服务商'}
      </h3>

      <div class="space-y-4">
        <div>
          <label for="provider-name" class="mb-1 block text-sm font-medium">名称</label>
          <input
            id="provider-name"
            type="text"
            bind:value={providerForm.name}
            placeholder="如：OpenAI 转发"
            class="w-full rounded-md border bg-background px-3 py-2"
          />
        </div>

        <div>
          <label for="provider-base" class="mb-1 block text-sm font-medium">API 地址</label>
          <input
            id="provider-base"
            type="text"
            bind:value={providerForm.openaiBase}
            placeholder="如：https://api.example.com/v1"
            class="w-full rounded-md border bg-background px-3 py-2"
          />
        </div>

        <div>
          <label for="provider-model" class="mb-1 block text-sm font-medium">模型列表</label>
          <div class="mb-2 flex gap-2">
            <input
              id="provider-model"
              type="text"
              bind:value={newModelInput}
              placeholder="输入模型名称"
              class="flex-1 rounded-md border bg-background px-3 py-2 text-sm"
              on:keydown={(e) => e.key === 'Enter' && addModel()}
            />
            <Button variant="outline" size="sm" on:click={addModel}>添加</Button>
          </div>

          {#if providerForm.models.length > 0}
            <div class="flex flex-wrap gap-2">
              {#each providerForm.models as model, i}
                <span class="inline-flex items-center gap-1 rounded bg-muted px-2 py-1 text-sm">
                  {model}
                  <button type="button" on:click={() => removeModel(i)} class="hover:text-destructive">×</button>
                </span>
              {/each}
            </div>
          {:else}
            <div class="text-sm text-muted-foreground">暂无模型</div>
          {/if}
        </div>
      </div>

      <div class="mt-6 flex gap-2">
        <Button variant="outline" on:click={closeProviderModal} className="flex-1">
          取消
        </Button>
        <Button variant="default" on:click={saveProvider} className="flex-1">
          保存
        </Button>
      </div>
    </div>
  </div>
{/if}
