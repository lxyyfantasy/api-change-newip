<template>
  <div>
    <el-card class="list-query" shadow="hover">
      <el-form inline label-width="60px">
        <el-form-item label="ID">
          <el-input v-model="listQuery.id" clearable/>
        </el-form-item>
        <el-form-item :label="T('Hostname')">
          <el-input v-model="listQuery.hostname" clearable/>
        </el-form-item>
        <el-form-item :label="T('LastOnlineTime')" label-width="100px">
          <el-select v-model="listQuery.time_ago" clearable>
            <el-option
                v-for="item in timeFilters"
                :key="item.value"
                :label="item.text"
                :value="item.value"
                :disabled="item.value === 0"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="T('Username')">
          <el-input v-model="listQuery.username" clearable/>
        </el-form-item>
        <el-form-item label="IP">
          <el-input v-model="listQuery.ip" clearable/>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handlerQuery">{{ T('Filter') }}</el-button>
          <el-button type="danger" @click="toAdd">{{ T('Add') }}</el-button>
          <el-button type="success" @click="toExport">{{ T('Export') }}</el-button>
          <el-popover :visible="showImport" placement="bottom" :width="600">
            <el-upload
                class="upload-demo"
                drag
                accept=".csv"
                :before-upload="parseCsv"
            >
              <el-icon class="el-icon--upload">
                <upload-filled/>
              </el-icon>
              <div class="el-upload__text">
                {{ T('Drop file here or click to upload') }}
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  {{ T('Please upload csv file') }} <br>
                  {{ T('Columns') }}: <span style="font-weight: bold;font-size: 15px">id,cpu,hostname,memory,os,username,uuid,version,group_id</span>
                  <br>
                  <span>{{ T('You can reference export file') }}</span>
                </div>
              </template>
            </el-upload>
            <el-button @click="showImport=false" type="primary">{{ T('Cancel') }}</el-button>
            <template #reference>
              <el-button @click="showImport=true" type="danger" :icon="ArrowDown">{{ T('Import') }}</el-button>
            </template>
          </el-popover>
          <el-button type="danger" @click="toBatchDelete">{{ T('BatchDelete') }}</el-button>
          <el-button type="primary" @click="toBatchAddToAB">{{ T('BatchAddToAB') }}</el-button>
          <el-button @click="openServerConfig">修改网络参数</el-button>
          <el-button @click="openTakeCodes">获取激活码</el-button>
          <el-button @click="openPending">待分配设备</el-button>
		  <el-button @click="openServerIp">设置服务器IP</el-button>
          <el-select :model-value="filterSelect" placeholder="筛选设备" style="width:160px;margin-left:8px" @change="onListFilterChange">
            <el-option label="全部设备" value="all" />
            <el-option label="在线设备" value="online" />
            <el-option label="离线设备" value="offline" />
            <el-option label="地址簿设备" value="owned" />
            <el-option label="被控端设备" value="orphan_server" />
            <el-option label="控制端设备" value="orphan_client" />
            <el-option label="服务器ip设备" value="server_ip_filter" />
            <el-option label="局域网ip设备" value="lan_ip_filter" />
          </el-select>
          <el-select v-model="tableSelect" placeholder="留存表" style="width:160px;margin-left:8px" @change="openTableBySelect">
            <el-option label="已删ID留存" value="removed_devices" />
            <el-option label="已删账号留存" value="removed_accounts" />
            <el-option label="已用开通码" value="used_lease" />
            <el-option label="已用托管码" value="used_host" />
            <el-option label="已分配设备" value="assigned" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card class="list-body" shadow="hover">
      <div style="text-align: right; margin-bottom: 10px">
        <el-button :icon="Setting" @click="showColumnSetting"></el-button>
      </div>

      <el-table :data="displayList" v-loading="listRes.loading" border size="small" height="calc(100vh - 300px)" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center"/>
        <template v-for="c in visibleColumns.filter(cc => cc.visible)" :key="c">
          <el-table-column v-if="c.name==='id'" prop="id" label="ID" align="center" width="150">
            <template #default="{row}">
              <span style="cursor:pointer;color:#409eff" @click="copyId(row.id)">{{ row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column v-if="c.name==='account'" prop="account" label="账号归属" align="center" width="120">
            <template #default="{row}">
              <span v-if="row.account === '__device__'">游离</span>
              <span v-else>{{ row.account }}</span>
            </template>
          </el-table-column>
          <el-table-column v-if="c.name==='category'" prop="category" label="分类" align="center" width="80"/>
		  <el-table-column v-if="c.name==='server_ip'" prop="server_ip" label="服务器ip" align="center" width="100"/>
		  <el-table-column v-if="c.name==='last_online_time'" prop="last_online_time" :label="T('LastOnlineTime')" align="center" min-width="120">
		    <template #default="{row}">
		      <div class="last_oline_time">
		        <span> {{ row.last_online_time ? timeAgo(row.last_online_time * 1000) : '-' }}</span> <span class="dot" :class="{red: timeDis(row.last_online_time) >= 60, green: timeDis(row.last_online_time)< 60}"></span>
		      </div>
		    </template>
		  </el-table-column>
          <el-table-column v-if="c.name==='connect_type'" prop="connect_type" label="连接" align="center" width="80">
            <template #default="{row}">
              <span v-if="row.connect_type==='direct'">P2P</span>
              <span v-else-if="row.connect_type==='relay'">中继</span>
              <span v-else-if="row.connect_type==='ip'">IP直连</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column v-if="c.name==='lan_ip'" prop="lan_ip" label="局域网ip" align="center" width="100"/>
          <el-table-column v-if="c.name==='lease_expire_at'" prop="lease_expire_at" label="开通有效期" align="center" width="150">
            <template #default="{row}">{{ fmtTime(row.lease_expire_at) }}</template>
          </el-table-column>
          <el-table-column v-if="c.name==='host_expire_at'" prop="host_expire_at" label="托管有效期" align="center" width="150">
            <template #default="{row}">{{ fmtTime(row.host_expire_at) }}</template>
          </el-table-column>
          <el-table-column v-if="c.name==='device_code'" prop="device_code" label="设备识别码" align="center" width="150"/>
          <el-table-column v-if="c.name==='lease_user'" prop="lease_user" label="开通激活账户" align="center" width="120"/>
          <el-table-column v-if="c.name==='lease_code'" prop="lease_code" label="开通激活码" align="center" width="150"/>
          <el-table-column v-if="c.name==='lease_days'" prop="lease_days" label="开通激活天数" align="center" width="100"/>
          <el-table-column v-if="c.name==='lease_at'" prop="lease_at" label="开通激活时间" align="center" width="150">
            <template #default="{row}">{{ fmtTime(row.lease_at) }}</template>
          </el-table-column>
          <el-table-column v-if="c.name==='host_user'" prop="host_user" label="托管激活账户" align="center" width="120"/>
          <el-table-column v-if="c.name==='host_code'" prop="host_code" label="托管激活码" align="center" width="150"/>
          <el-table-column v-if="c.name==='host_days'" prop="host_days" label="托管激活天数" align="center" width="100"/>
          <el-table-column v-if="c.name==='host_at'" prop="host_at" label="托管激活时间" align="center" width="150">
            <template #default="{row}">{{ fmtTime(row.host_at) }}</template>
          </el-table-column>
          <el-table-column v-if="c.name==='cpu'" prop="cpu" label="CPU" align="center" width="100" show-overflow-tooltip/>
          <el-table-column v-if="c.name==='hostname'" prop="hostname" :label="T('Hostname')" align="center" width="120"/>
          <el-table-column v-if="c.name==='memory'" prop="memory" :label="T('Memory')" align="center" width="120"/>
          <el-table-column v-if="c.name==='os'" prop="os" :label="T('Os')" align="center" width="120" show-overflow-tooltip/>
          <el-table-column v-if="c.name==='last_online_ip'" prop="last_online_ip" :label="T('LastOnlineIp')" align="center" min-width="120"/>
          <el-table-column v-if="c.name==='username'" prop="username" :label="T('Username')" align="center" width="120"/>
          <el-table-column v-if="c.name==='group_id'" prop="group_id" :label="T('Group')" align="center" width="120">
            <template #default="{row}">
              <span v-if="row.group_id"> <el-tag>{{ groupListRes.list?.find(g => g.id === row.group_id)?.name }} </el-tag> </span>
              <span v-else> - </span>
            </template>
          </el-table-column>
          <el-table-column v-if="c.name==='uuid'" prop="uuid" :label="T('Uuid')" align="center" width="120" show-overflow-tooltip/>
          <el-table-column v-if="c.name==='version'" prop="version" :label="T('Version')" align="center" width="80"/>
          <el-table-column v-if="c.name==='alias'" prop="alias" :label="T('Alias')" align="center" width="80"/>
          <el-table-column v-if="c.name==='created_at'" prop="created_at" :label="T('CreatedAt')" align="center" width="150"/>
          <el-table-column v-if="c.name==='updated_at'" prop="updated_at" :label="T('UpdatedAt')" align="center" width="150"/>
        </template>

        <el-table-column :label="T('Actions')" align="center" width="500" class-name="table-actions" fixed="right">
          <template #default="{row}">
            <el-button v-if="appStore.setting.appConfig.web_client" type="success" @click="toWebClientLink(row)">Web Client</el-button>
            <el-button type="primary" @click="toAddressBook(row)">{{ T('AddToAddressBook') }}</el-button>
            <el-button @click="toEdit(row)">{{ T('Edit') }}</el-button>
            <el-button type="danger" @click="del(row)">{{ T('Delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="formVisible" :title="!formData.row_id?T('Create'):T('Update')" width="800">
      <el-form class="dialog-form" ref="form" :model="formData" label-width="120px">
        <el-form-item label="ID" prop="id" required>
          <el-input v-model="formData.id"></el-input>
        </el-form-item>
        <el-form-item :label="T('Group')" prop="group_id">
          <el-select v-model="formData.group_id">
            <el-option
                v-for="item in groupListRes.list"
                :key="item.id"
                :label="item.name"
                :value="item.id"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="T('Username')" prop="username">
          <el-input v-model="formData.username"></el-input>
        </el-form-item>
        <el-form-item :label="T('Hostname')" prop="hostname">
          <el-input v-model="formData.hostname"></el-input>
        </el-form-item>
        <el-form-item label="CPU" prop="cpu">
          <el-input v-model="formData.cpu"></el-input>
        </el-form-item>
        <el-form-item :label="T('Memory')" prop="memory">
          <el-input v-model="formData.memory"></el-input>
        </el-form-item>
        <el-form-item :label="T('Os')" prop="os">
          <el-input v-model="formData.os"></el-input>
        </el-form-item>
        <el-form-item :label="T('Uuid')" prop="uuid">
          <el-input v-model="formData.uuid"></el-input>
        </el-form-item>
        <el-form-item :label="T('Version')" prop="version">
          <el-input v-model="formData.version"></el-input>
        </el-form-item>
        <el-form-item :label="T('Alias')" prop="alias">
          <el-input v-model="formData.alias"></el-input>
        </el-form-item>
		<el-form-item label="服务器" prop="server_ip">
		  <el-input v-model="formData.server_ip"></el-input>
		</el-form-item>
        <el-form-item>
          <el-button @click="formVisible = false">{{ T('Cancel') }}</el-button>
          <el-button @click="submit" type="primary">{{ T('Submit') }}</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <el-dialog v-model="ABFormVisible" width="800" :title="T('Create')" destroy-on-close>
      <createABForm :peer="clickRow" @success="ABFormVisible=false" @cancel="ABFormVisible=false"></createABForm>
    </el-dialog>

    <el-dialog v-model="batchABFormVisible" width="800" :title="T('Create')">
      <el-form class="dialog-form" ref="form" :model="batchABFormData" label-width="120px">
        <el-form-item :label="T('Owner')" prop="user_id" required>
          <el-select v-model="batchABFormData.user_id" @change="changeUserForBatchCreateAB">
            <el-option
                v-for="item in allUsers"
                :key="item.id"
                :label="item.username"
                :value="item.id"
            ></el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="T('AddressBookName')" required prop="collection_id">
          <el-select v-model="batchABFormData.collection_id" clearable>
            <el-option :value="0" :label="T('MyAddressBook')"></el-option>
            <el-option v-for="c in collectionListResForBatchCreateAB.list" :key="c.id" :label="c.name" :value="c.id"></el-option>
          </el-select>
        </el-form-item>
        <!--        <el-form-item :label="T('Tags')" prop="tags">
                  <el-select v-model="batchABFormData.tags" multiple>
                    <el-option
                        v-for="item in tagListRes.list"
                        :key="item.name"
                        :label="item.name"
                        :value="item.name"
                    ></el-option>
                  </el-select>
                </el-form-item>-->
        <el-form-item>
          <el-button @click="batchABFormVisible = false">{{ T('Cancel') }}</el-button>
          <el-button @click="submitBatchAddToAB" type="primary">{{ T('Submit') }}</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <el-dialog v-model="columnSettingVisible" title="Column Setting">
      <div v-for="(row, key) in visibleColumns" :key="key" style="margin-bottom: 10px;display: flex;align-items: center">
        <div style="width: 200px">
          <el-checkbox v-model="row.visible" :label="true">{{ T(row.label) }}</el-checkbox>
        </div>
        <div @click="upColumn(key)" style="width: 100px;cursor: pointer">
          <el-icon :size="20">
            <ArrowUp/>
          </el-icon>
        </div>
        <div @click="downColumn(key)" style="width: 100px;cursor: pointer">
          <el-icon :size="20">
            <ArrowDown/>
          </el-icon>
        </div>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button @click="columnSettingVisible = false">{{ T('Cancel') }}</el-button>
        <el-button type="primary" @click="saveColumnSetting">{{ T('Save') }}</el-button>
      </span>
    </el-dialog>
    <el-dialog v-model="serverConfigVisible" title="修改网络参数" width="500">
      <el-form label-width="100px">
        <el-form-item label="Key"><el-input v-model="serverConfig.server_key" /></el-form-item>
        <el-form-item label="ip1"><el-input v-model="serverConfig.ip1" /></el-form-item>
        <el-form-item label="ip2"><el-input v-model="serverConfig.ip2" /></el-form-item>
        <el-form-item label="管理员密码"><el-input v-model="serverConfig.password" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="serverConfigVisible = false">取消</el-button>
        <el-button type="primary" @click="saveServerConfig">保存</el-button>
      </template>
    </el-dialog>
    
    <el-dialog v-model="takeCodesVisible" title="获取激活码" width="520">
      <el-form label-width="90px">
        <el-form-item label="类别">
          <el-select v-model="takeCodesForm.kind" style="width:100%">
            <el-option label="开通" value="lease" />
            <el-option label="托管" value="host" />
            <el-option label="邀请码" value="invite" />
            <el-option label="添加设备码" value="adddevice" />
          </el-select>
        </el-form-item>
        <el-form-item label="天数" v-if="takeCodesForm.kind==='lease' || takeCodesForm.kind==='host'">
          <el-select v-model="takeCodesForm.days" style="width:100%">
            <el-option :value="7" label="7" />
            <el-option :value="15" label="15" />
            <el-option :value="30" label="30" />
          </el-select>
        </el-form-item>
        <el-form-item label="数量">
          <el-select v-model="takeCodesForm.n" style="width:100%">
            <el-option v-for="i in 20" :key="i" :value="i" :label="i" />
          </el-select>
        </el-form-item>
        <el-form-item label="管理员密码">
          <el-input v-model="takeCodesForm.password" type="password" show-password />
        </el-form-item>
      </el-form>
      <div v-if="takeCodesForm.result.length" style="margin-top:10px">
        <el-input type="textarea" :rows="8" :model-value="takeCodesForm.result.join('\n')" readonly />
        <el-button size="small" style="margin-top:6px" @click="copyCode(takeCodesForm.result.join('\n'))">复制全部</el-button>
      </div>
      <template #footer>
        <el-button @click="takeCodesVisible = false">关闭</el-button>
        <el-button type="primary" @click="doTakeCodes">获取</el-button>
      </template>
    </el-dialog>
    
    <el-dialog v-model="tableDialogVisible" :title="tableDialogTitle" width="900">
      <el-table :data="tableDialogData" border size="small" height="400">
        <el-table-column v-for="col in tableDialogCols" :key="col.prop"
                         :prop="col.prop" :label="col.label" align="center" />
      </el-table>
    </el-dialog>
    <el-dialog v-model="pendingVisible" title="待分配设备" width="520">
      <el-form label-width="40px" inline>
        <el-form-item label="ID" style="margin-bottom:0">
          <el-input v-model="pendingForm.id" placeholder="输入ID" style="width:220px" />
        </el-form-item>
        <el-form-item style="margin-bottom:0">
          <el-button type="primary" @click="doPendingAdd">添加</el-button>
        </el-form-item>
      </el-form>
        <el-table :data="pendingListData" border size="small" height="300">
        <el-table-column prop="id" label="ID" align="center" />
        <el-table-column label="操作" align="center" width="100">
          <template #default="{row}">
            <el-button size="small" type="danger" @click="doPendingDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="pendingVisible = false">关闭</el-button>
      </template>
    </el-dialog>
	<el-dialog v-model="serverIpVisible" title="设置服务器IP" width="520">
	  <el-form label-width="90px">
	    <el-form-item label="设备ID">
	      <el-input v-model="serverIpForm.id" placeholder="输入设备ID"/>
	    </el-form-item>
	    <el-form-item label="服务器IP">
	      <el-select v-model="serverIpForm.server_ip" placeholder="选择服务器IP，可留空清除" clearable style="width:100%">
	        <el-option v-for="s in serverInfoList" :key="s.id" :label="s.ip" :value="s.ip" />
	      </el-select>
	    </el-form-item>
	  </el-form>

	  <el-divider content-position="left">添加服务器</el-divider>
  	  <div style="display:flex;justify-content:space-between;gap:10px">
	    <el-input v-model="serverAddForm.ip" placeholder="IP" style="flex:1" />
	    <el-input v-model="serverAddForm.bandwidth" placeholder="带宽" style="flex:1" />
	    <el-button type="primary" @click="doServerAdd" style="width:80px">添加</el-button>
	  </div>

	  <el-divider content-position="left">服务器列表</el-divider>
	  <el-table :data="serverInfoList" border size="small" height="220">
	    <el-table-column prop="ip" label="IP" align="center" />
	    <el-table-column prop="bandwidth" label="带宽" align="center" />
	    <el-table-column prop="device_count" label="设备数量" align="center" width="100" />
	    <el-table-column label="操作" align="center" width="90">
	      <template #default="{row}">
	        <el-button size="small" type="danger" @click="doServerDelete(row)">删除</el-button>
	      </template>
	    </el-table-column>
	  </el-table>

	  <template #footer>
	    <el-button @click="serverIpVisible = false">取消</el-button>
	    <el-button type="primary" @click="doSetServerIp">确定</el-button>
	  </template>
	</el-dialog>
	<el-dialog v-model="ipFilterVisible" :title="ipFilterForm.type === 'server_ip_filter' ? '按服务器IP筛选' : '按局域网IP筛选'" width="420">
	  <el-form label-width="90px">
	    <el-form-item :label="ipFilterForm.type === 'server_ip_filter' ? '服务器IP' : '局域网IP'">
	      <el-input v-model="ipFilterForm.ip" placeholder="输入ip" clearable/>
	    </el-form-item>
	  </el-form>
	  <template #footer>
	    <el-button @click="cancelIpFilter">取消</el-button>
	    <el-button type="primary" @click="confirmIpFilter">确定</el-button>
	  </template>
	</el-dialog>
  </div>
</template>

<script setup>
  import { computed, onActivated, onMounted, reactive, ref, watch } from 'vue'
  import { all, batchRemove, create, list, remove, update, serverConfigGet, serverConfigPut, takeCodes, removedDevices, removedAccounts, usedCodes, pendingList, pendingAdd, pendingDelete, assignedList, setServerIp, serverList, serverCreate, serverDelete } from '@/api/peer'
  import { list as groupList } from '@/api/device_group'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { toWebClientLink } from '@/utils/webclient'
  import { T } from '@/utils/i18n'
  import { timeAgo } from '@/utils/time'
  import { jsonToCsv, downBlob } from '@/utils/file'
  import { loadAllUsers } from '@/global'
  import { useAppStore } from '@/store/app'
  import { connectByClient } from '@/utils/peer'
  import { ArrowDown, ArrowUp, CopyDocument, Setting } from '@element-plus/icons'
  import { handleClipboard } from '@/utils/clipboard'
  import { batchCreateFromPeers } from '@/api/address_book'
  import { useRepositories as useCollectionRepositories } from '@/views/address_book/collection'
  import createABForm from '@/views/peer/createABForm.vue'
  import { UploadFilled } from '@element-plus/icons-vue'

  const appStore = useAppStore()

  //group
  const groupListRes = reactive({
    list: [], total: 0, loading: false,
  })
  const groupListQuery = reactive({
    page: 1,
    page_size: 999,
  })
  const getGroupList = async () => {
    groupListRes.loading = true
    const res = await groupList(groupListQuery).catch(_ => false)
    groupListRes.loading = false
    if (res) {
      groupListRes.list = res.data.list
      groupListRes.total = res.data.total
    }
  }
  onMounted(getGroupList)
  //

  const listRes = reactive({
    list: [], total: 0, loading: false,
  })
  const displayList = ref([])
  const listFilter = ref('all')
  const tableSelect = ref('')
  
  // 修改网络参数
  const serverConfigVisible = ref(false)
  const serverConfig = reactive({ server_key: '', ip1: '', ip2: '', password: '' })
  const openServerConfig = async () => {
    serverConfigVisible.value = true
    serverConfig.password = ''
    const res = await serverConfigGet().catch(_ => false)
    if (res) {
      serverConfig.server_key = res.server_key || ''
      serverConfig.ip1 = res.ip1 || ''
      serverConfig.ip2 = res.ip2 || ''
    }
  }
  const saveServerConfig = async () => {
    if (!serverConfig.password) {
      ElMessage.warning('请输入管理员密码')
      return
    }
    const res = await serverConfigPut({
      key: serverConfig.server_key,
      ip1: serverConfig.ip1,
      ip2: serverConfig.ip2,
      password: serverConfig.password,
    }).catch(_ => false)
    if (res) {
      ElMessage.success('保存成功')
      serverConfigVisible.value = false
    }
  }
  
  // 获取激活码
  const takeCodesVisible = ref(false)
  const takeCodesForm = reactive({ kind: 'lease', days: 7, n: 1, password: '', result: [] })
  const openTakeCodes = () => {
    takeCodesVisible.value = true
    takeCodesForm.result = []
    takeCodesForm.password = ''
  }
  const doTakeCodes = async () => {
    if (!takeCodesForm.password) {
      ElMessage.warning('请输入管理员密码')
      return
    }
    const params = {
      kind: takeCodesForm.kind,
      n: takeCodesForm.n,
      password: takeCodesForm.password,
    }
    if (takeCodesForm.kind === 'lease' || takeCodesForm.kind === 'host') {
      params.days = takeCodesForm.days
    }
    const res = await takeCodes({
      ...params,
    }).catch(_ => false)
    if (res) {
      takeCodesForm.result = res.data.list || []
    }
  }
  const copyCode = async (text) => {
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success('已复制')
    } catch (e) {
      ElMessage.warning('复制失败，请手动复制')
    }
  }

    // 待分配设备
  const pendingVisible = ref(false)
  const pendingForm = reactive({ id: '' })
  const pendingListData = ref([])
  const loadPending = async () => {
    const res = await pendingList().catch(_ => false)
    if (res) {
      pendingListData.value = res.data.list || []
    }
  }
  const openPending = () => {
    pendingVisible.value = true
    pendingForm.id = ''
    loadPending()
  }
  const doPendingAdd = async () => {
    if (!pendingForm.id) {
      ElMessage.warning('请输入ID')
      return
    }
    const pwd = await ElMessageBox.prompt('请输入管理员密码', '确认', {
      inputType: 'password',
    }).catch(_ => false)
    if (!pwd) return
    const res = await pendingAdd({
      id: pendingForm.id,
      password: pwd.value,
    }).catch(_ => false)
    if (res) {
      ElMessage.success('添加成功')
      pendingForm.id = ''
      loadPending()
    }
  }
  const doPendingDelete = async (row) => {
    const pwd = await ElMessageBox.prompt('请输入管理员密码', '确认删除', {
      inputType: 'password',
    }).catch(_ => false)
    if (!pwd) return
    const res = await pendingDelete({ id: row.id, password: pwd.value }).catch(_ => false)
    if (res) {
      ElMessage.success('删除成功')
      loadPending()
    }
  }

  // 留存表弹窗
  const tableDialogVisible = ref(false)
  const tableDialogTitle = ref('')
  const tableDialogData = ref([])
  const tableDialogCols = ref([])
  const openTableBySelect = async (val) => {
    if (!val) return
    if (val === 'removed_devices') {
      const res = await removedDevices().catch(_ => false)
      if (res) {
        tableDialogTitle.value = '已删除ID留存'
        tableDialogCols.value = [
          { prop: 'id', label: 'ID' },
          { prop: 'account', label: '账号归属' },
          { prop: 'removed_at', label: '删除时间' },
        ]
        tableDialogData.value = res.data.list || []
        tableDialogVisible.value = true
      }
    } else if (val === 'removed_accounts') {
      const res = await removedAccounts().catch(_ => false)
      if (res) {
        tableDialogTitle.value = '已删除账号留存'
        tableDialogCols.value = [
          { prop: 'account', label: '账号' },
          { prop: 'empty_since', label: '开始计时' },
        ]
        tableDialogData.value = res.data.list || []
        tableDialogVisible.value = true
      }
    } else if (val === 'assigned') {
      const res = await assignedList().catch(_ => false)
      if (res) {
        tableDialogTitle.value = '已分配设备'
        tableDialogCols.value = [
          { prop: 'id', label: 'ID' },
          { prop: 'account', label: '分配账号' },
          { prop: 'assigned_at', label: '分配时间' },
        ]
        tableDialogData.value = (res.data.list || []).map(r => ({
          ...r,
          assigned_at: fmtTime(r.assigned_at),
        }))
        tableDialogVisible.value = true
      }
    } else {
      const kind = val === 'used_lease' ? 'lease' : 'host'
      const res = await usedCodes({ kind }).catch(_ => false)
      if (res) {
        tableDialogTitle.value = kind === 'lease' ? '已用开通激活码' : '已用托管激活码'
        tableDialogCols.value = [
          { prop: 'code', label: '激活码' },
          { prop: 'days', label: '天数' },
          { prop: 'bind_id', label: '绑定ID' },
          { prop: 'used_at', label: '使用时间' },
        ]
        tableDialogData.value = res.data.list || []
        tableDialogVisible.value = true
      }
    }
    tableSelect.value = ''
  }

  const listQuery = reactive({
    page: 1,
    page_size: 100000,
    time_ago: null,
    id: '',
    hostname: '',
    username: '',
    ip: '',
  })

  const getList = async () => {
    listRes.loading = true
    const res = await all().catch(_ => false)
    listRes.loading = false
    if (res) {
      listRes.list = res.list || []
      listRes.total = listRes.list.length
      applyFilter()
    }
  }
  const applyFilter = () => {
    const q = listQuery
    let arr = listRes.list
    if (q.id) arr = arr.filter(r => String(r.id).includes(q.id))
    if (q.hostname) arr = arr.filter(r => (r.hostname || '').includes(q.hostname))
    if (q.username) arr = arr.filter(r => (r.username || '').includes(q.username))
    if (q.ip) arr = arr.filter(r => (r.last_online_ip || '').includes(q.ip))
    if (q.time_ago !== null && q.time_ago !== undefined && q.time_ago !== '') {
      const now = Math.floor(Date.now() / 1000)
      if (q.time_ago < 0) {
        // 小于 N 时间
        arr = arr.filter(r => r.last_online_time && (now - r.last_online_time) < Math.abs(q.time_ago))
      } else {
        // 大于 N 时间前
        arr = arr.filter(r => r.last_online_time && (now - r.last_online_time) > q.time_ago)
      }
    }
    if (listFilter.value === 'online') {
      const now = Math.floor(Date.now() / 1000)
      arr = arr.filter(r => r.last_online_time && (now - r.last_online_time) < 60)
    } else if (listFilter.value === 'offline') {
      const now = Math.floor(Date.now() / 1000)
      arr = arr.filter(r => !r.last_online_time || (now - r.last_online_time) >= 60)
    } else if (listFilter.value === 'owned') {
      arr = arr.filter(r => r.account && r.account !== '__device__')
    } else if (listFilter.value === 'orphan_server') {
      arr = arr.filter(r => r.category === '被控端')
    } else if (listFilter.value === 'orphan_client') {
      arr = arr.filter(r => r.category === '控制端')
    } else if (listFilter.value === 'server_ip_filter') {
      if (ipFilterForm.ip) arr = arr.filter(r => (r.server_ip || '').includes(ipFilterForm.ip))
    } else if (listFilter.value === 'lan_ip_filter') {
      if (ipFilterForm.ip) arr = arr.filter(r => (r.lan_ip || '').includes(ipFilterForm.ip))
    }
    displayList.value = arr
  }

  // ip 筛选弹窗
  const ipFilterVisible = ref(false)
  const ipFilterForm = reactive({ type: '', ip: '' })

  const filterSelect = ref('')
  const openIpFilter = (type) => {
    ipFilterForm.type = type
    ipFilterForm.ip = ''
    ipFilterVisible.value = true
  }
  const onListFilterChange = (val) => {
    if (val === 'server_ip_filter' || val === 'lan_ip_filter') {
      listFilter.value = val
      openIpFilter(val)
      filterSelect.value = ''
      return
    }
    filterSelect.value = val
    listFilter.value = val
    applyFilter()
  }
  const confirmIpFilter = () => {
    ipFilterVisible.value = false
    filterSelect.value = ''
    applyFilter()
  }
  const cancelIpFilter = () => {
    ipFilterVisible.value = false
    filterSelect.value = ''
    listFilter.value = 'all'
    applyFilter()
  }

  const applyListFilter = () => {
    applyFilter()
  }

  const handlerQuery = () => {
    applyFilter()
  }

  const del = async (row) => {
    const cf = await ElMessageBox.confirm(T('Confirm?', { param: T('Delete') }), {
      confirmButtonText: T('Confirm'),
      cancelButtonText: T('Cancel'),
      type: 'warning',
    }).catch(_ => false)
    if (!cf) {
      return false
    }

    const res = await remove({ row_id: row.row_id }).catch(_ => false)
    if (res) {
      ElMessage.success(T('OperationSuccess'))
      getList()
    }
  }
  onMounted(getList)
  onActivated(getList)

  // watch(() => listQuery.page, getList)

  // watch(() => listQuery.page_size, handlerQuery)

  const formVisible = ref(false)
  const formData = reactive({
    row_id: 0,
    group_id: null,
    cpu: '',
    hostname: '',
    id: '',
    memory: '',
    os: '',
    username: '',
    uuid: '',
    version: '',
	server_ip: '',
  })

  const toEdit = (row) => {
    formVisible.value = true
    //将row中的数据赋值给formData
    Object.keys(formData).forEach(key => {
      formData[key] = row[key]
    })
  }
  const toAdd = () => {
    formVisible.value = true
    //重置formData
    formData.row_id = 0
    formData.cpu = ''
    formData.hostname = ''
    formData.id = ''
    formData.memory = ''
    formData.os = ''
    formData.username = ''
    formData.uuid = ''
    formData.version = ''
	formData.server_ip = ''
  }
  const submit = async () => {
    const api = formData.row_id ? update : create
    const res = await api(formData).catch(_ => false)
    if (res) {
      ElMessage.success(T('OperationSuccess'))
      formVisible.value = false
      getList()
    }
  }

  const copyId = async (id) => {
    try {
      await navigator.clipboard.writeText(id)
      ElMessage({ message: '已复制', type: 'success', duration: 1000 })
    } catch (e) {
      // 复制失败不提示
    }
  }

  const fmtTime = (ts) => {
    if (!ts || ts <= 0) return ''
    const d = new Date(ts * 1000)
    const p = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${p(d.getMonth()+1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
  }

  const timeDis = (time) => {
    let now = new Date().getTime()
    let after = new Date(time * 1000).getTime()
    return (now - after) / 1000
  }

  const timeFilters = computed(() => [
    { text: T('MinutesLess', { param: 1 }, 1), value: -60 },
    { text: T('HoursLess', { param: 1 }, 1), value: -3600 },
    { text: T('DaysLess', { param: 1 }, 1), value: -86400 },
    { text: '---------', value: 0 },
    { text: T('MinutesAgo', { param: 1 }, 1), value: 60 },
    { text: T('HoursAgo', { param: 1 }, 1), value: 3600 },
    { text: T('DaysAgo', { param: 1 }, 1), value: 86400 },
    { text: T('MonthsAgo', { param: 1 }, 1), value: 2592000 },
    // { text: T('YearsAgo', { param: 1 }, 1), value: 31536000 },
  ])

  const toExport = async () => {
    const data = displayList.value.map(item => {
      const o = { ...item }
      o.last_online_time = o.last_online_time ? new Date(o.last_online_time * 1000).toLocaleString() : '-'
      delete o.user_id
      delete o.user
      return o
    })
    const csv = jsonToCsv(data)
    downBlob(csv, 'peers.csv')
  }

  const showImport = ref(false)
  const canKeys = ['id', 'cpu', 'hostname', 'memory', 'os', 'username', 'uuid', 'version', 'group_id']
  const parseCsv = (file) => {
    const reader = new FileReader()
    reader.onload = async (e) => {
      const data = e.target.result
      console.log(data)
      //组装数据
      const rows = data.split('\n')
      const keys = rows[0].split(',')
      console.log(keys, rows.slice(1).map(row => row.split(',')))
      const values = rows.slice(1).map(row => {
        const obj = {}
        row.split(/,(?=(?:(?:[^"]*"){2})*[^"]*$)/).forEach((v, i) => {
          //去掉两边的"
          obj[keys[i]] = v.trim().replace(/^"|"$/g, '')
        })
        return obj
      }).filter(item => item.id)
      // console.log(values)
      //移除不需要的key
      values.forEach(item => {
        item.group_id = parseInt(item.group_id)
        Object.keys(item).forEach(key => {
          if (!canKeys.includes(key)) {
            delete item[key]
          }
        })
      })
      console.log(values)
      const pa = []
      values.map(item => {
        pa.push(create(item))
      })
      const res = await Promise.all(pa).catch(_ => false)
      if (res) {
        ElMessage.success(T('OperationSuccess'))
        getList()
      }

    }
    reader.readAsText(file)
    return false
  }
  const toImport = () => {
    ElMessage.warning('暂未实现')
  }

  const ABFormVisible = ref(false)
  const clickRow = ref({})
  const toAddressBook = (row) => {
    clickRow.value = row
    ABFormVisible.value = true
  }

  const multipleSelection = ref([])
  const handleSelectionChange = (val) => {
    multipleSelection.value = val
  }
  const toBatchDelete = async () => {
    if (!multipleSelection.value.length) {
      ElMessage.warning(T('PleaseSelectData'))
      return false
    }
    const cf = await ElMessageBox.confirm(T('Confirm?', { param: T('BatchDelete') }), {
      confirmButtonText: T('Confirm'),
      cancelButtonText: T('Cancel'),
      type: 'warning',
    }).catch(_ => false)
    if (!cf) {
      return false
    }

    const res = await batchRemove({ row_ids: multipleSelection.value.map(i => i.row_id) }).catch(_ => false)
    if (res) {
      ElMessage.success(T('OperationSuccess'))
      getList()
    }
  }

  // 批量添加到地址簿 start
  const { allUsers, getAllUsers } = loadAllUsers()
  onMounted(getAllUsers)
  const {
    listRes: collectionListResForBatchCreateAB,
    listQuery: collectionListQueryForBatchCreateAB,
    getList: getCollectionListForBatchCreateAB,
  } = useCollectionRepositories('admin')
  collectionListQueryForBatchCreateAB.page_size = 9999
  const changeUserForBatchCreateAB = (val) => {
    batchABFormData.value.collection_id = 0
    collectionListQueryForBatchCreateAB.user_id = val
    getCollectionListForBatchCreateAB()
  }
  const batchABFormVisible = ref(false)
  const toBatchAddToAB = () => {
    batchABFormVisible.value = true
  }
  const batchABFormData = ref({
    collection_id: 0,
    tags: [],
    peer_ids: [],
    user_id: null,
  })
  const submitBatchAddToAB = async () => {
    if (multipleSelection.value.length === 0) {
      ElMessage.warning(T('PleaseSelectData'))
      return false
    }
    batchABFormData.value.peer_ids = multipleSelection.value.map(i => i.row_id)
    if (!batchABFormData.value.peer_ids.length) {
      ElMessage.warning(T('PleaseSelectData'))
      return false
    }

    const res = await batchCreateFromPeers(batchABFormData.value).catch(_ => false)
    if (res) {
      ElMessage.success(T('OperationSuccess'))
      batchABFormVisible.value = false
    }
  }
  // 批量添加到地址簿 end

  const columnSettingVisible = ref(false)
  const allColumns = ref([
    { name: 'id', visible: true, label: 'Id' },
    // ===== 新增 14 列 =====
    { name: 'account', visible: true, label: '账号归属' },
    { name: 'category', visible: true, label: '分类' },
	{ name: 'server_ip', visible: true, label: '服务器' },
    { name: 'last_online_time', visible: true, label: 'LastOnlineTime' },
    { name: 'connect_type', visible: true, label: '连接' },
    { name: 'lan_ip', visible: true, label: '局域网ip' },
    { name: 'lease_expire_at', visible: true, label: '开通有效期' },
    { name: 'host_expire_at', visible: true, label: '托管有效期' },
    { name: 'device_code', visible: true, label: '设备识别码' },
    { name: 'lease_user', visible: true, label: '开通激活账户' },
    { name: 'lease_code', visible: true, label: '开通激活码' },
    { name: 'lease_days', visible: true, label: '开通激活天数' },
    { name: 'lease_at', visible: true, label: '开通激活时间' },
    { name: 'host_user', visible: true, label: '托管激活账户' },
    { name: 'host_code', visible: true, label: '托管激活码' },
    { name: 'host_days', visible: true, label: '托管激活天数' },
    { name: 'host_at', visible: true, label: '托管激活时间' },
    // ===== 原有 =====
    { name: 'cpu', visible: true, label: 'Cpu' },
    { name: 'hostname', visible: true, label: 'Hostname' },
    { name: 'memory', visible: true, label: 'Memory' },
    { name: 'os', visible: true, label: 'Os' },
    { name: 'last_online_ip', visible: true, label: 'LastOnlineIp' },
    { name: 'username', visible: true, label: 'Username' },
    { name: 'group_id', visible: true, label: 'Group' },
    { name: 'uuid', visible: true, label: 'Uuid' },
    { name: 'version', visible: true, label: 'Version' },
    { name: 'alias', visible: true, label: 'Alias' },
    { name: 'created_at', visible: true, label: 'CreatedAt' },
    { name: 'updated_at', visible: true, label: 'UpdatedAt' },
  ])
  const visibleColumns = ref(JSON.parse(localStorage.getItem('peer_visible_columns')) || allColumns.value)
  const showColumnSetting = () => {
    columnSettingVisible.value = true
  }
  const saveColumnSetting = () => {
    localStorage.setItem('peer_visible_columns', JSON.stringify(visibleColumns.value))
    ElMessage.success(T('OperationSuccess'))
    columnSettingVisible.value = false
  }

  const upColumn = (index) => {
    if (index === 0) return
    const col = visibleColumns.value[index]
    visibleColumns.value.splice(index, 1)
    visibleColumns.value.splice(index - 1, 0, col)

  }
  const downColumn = (index) => {
    if (index === visibleColumns.value.length - 1) return
    const col = visibleColumns.value[index]
    visibleColumns.value.splice(index, 1)
    visibleColumns.value.splice(index + 1, 0, col)

  }

  const serverIpVisible = ref(false)
  const serverIpForm = reactive({ id: '', server_ip: '' })
  const serverInfoList = ref([])
  const serverAddForm = reactive({ ip: '', bandwidth: '' })

  const loadServerInfoList = async () => {
    const res = await serverList().catch(_ => false)
    if (res) {
      serverInfoList.value = res.data.list || []
    }
  }
  const openServerIp = () => {
    serverIpVisible.value = true
    loadServerInfoList()
  }
  const doSetServerIp = async () => {
    if (!serverIpForm.id) {
      ElMessage.warning('请输入设备ID')
      return
    }
    const res = await setServerIp({ ...serverIpForm }).catch(_ => false)
    if (res) {
      ElMessage.success('设置成功')
      serverIpVisible.value = false
      serverIpForm.id = ''
      serverIpForm.server_ip = ''
      getList()
    }
  }
  const doServerAdd = async () => {
    if (!serverAddForm.ip) {
      ElMessage.warning('请输入IP')
      return
    }
    const res = await serverCreate({
      ip: serverAddForm.ip,
      bandwidth: serverAddForm.bandwidth,
    }).catch(_ => false)
    if (res) {
      ElMessage.success('添加成功')
      serverAddForm.ip = ''
      serverAddForm.bandwidth = ''
      loadServerInfoList()
    }
  }
  const doServerDelete = async (row) => {
    const res = await serverDelete({ id: row.id }).catch(_ => false)
    if (res) {
      ElMessage.success('删除成功')
      loadServerInfoList()
    }
  }
</script>

<style scoped lang="scss">
.list-body {
  :deep(.el-scrollbar__bar.is-horizontal) {
    height: 12px;
    bottom: 2px;
  }
  :deep(.el-scrollbar__bar.is-horizontal .el-scrollbar__thumb) {
    height: 100%;
    opacity: 0.5;
  }
}
.list-query .el-select {
  --el-select-width: 180px;
}

.last_oline_time {
  display: flex;
  justify-content: center;
  align-items: center;
}

.dot {
  width: 6px;
  height: 6px;
  display: block;
  border-radius: 50%;
  margin-left: 10px;

  &.red {
    background-color: red;
  }

  &.green {
    background-color: green;
  }
}
</style>
