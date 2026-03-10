<template>
  <div class="app">
    <h1>文件备份系统</h1>
    
    <!-- 登录弹窗 -->
    <div v-if="!isLoggedIn" class="modal login-modal">
      <div class="modal-content">
        <h2>请登录</h2>
        <div class="form-group">
          <label for="password">密码：</label>
          <input 
            type="password" 
            id="password" 
            v-model="loginPassword" 
            placeholder="请输入密码"
            @keyup.enter="login"
          />
        </div>
        <div class="modal-footer">
          <button class="btn" @click="login">登录</button>
        </div>
        <p v-if="loginError" class="error-message">{{ loginError }}</p>
      </div>
    </div>
    
    <div v-if="isLoggedIn">
      <div class="tabs">
        <button 
          v-for="tab in tabs" 
          :key="tab.id"
          :class="{ active: activeTab === tab.id }"
          @click="activeTab = tab.id"
        >
          {{ tab.name }}
        </button>
        <button class="logout-btn" @click="logout">退出登录</button>
      </div>

    <div class="tab-content">
      <!-- 备份标签页 -->
      <div v-if="activeTab === 'backup'" class="tab-pane">
        <h2>创建备份</h2>
        <form @submit.prevent="createBackup">
          <div class="form-group">
            <label for="sourceDir">源目录:</label>
            <div class="input-with-button">
              <input 
                type="text" 
                id="sourceDir" 
                v-model="backupForm.sourceDir"
                placeholder="请选择要备份的目录"
                required
              />
              <button type="button" class="btn select-btn" @click="openSourceDirDialog">
                选择
              </button>
            </div>
          </div>
          <div class="form-group">
            <label for="outputDir">输出目录:</label>
            <input 
              type="text" 
              id="outputDir" 
              v-model="backupForm.outputDir"
              placeholder="备份文件保存路径"
              readonly
            />
          </div>
          <div class="form-group">
            <label for="fileName">文件名 (可选):</label>
            <input 
              type="text" 
              id="fileName" 
              v-model="backupForm.fileName"
              placeholder="不填则自动生成"
            />
          </div>
          <button type="submit" class="btn" :disabled="isLoading">
            {{ isLoading ? '备份中...' : '开始备份' }}
          </button>
        </form>
        <div v-if="backupResult" class="result">
          <h3>备份结果:</h3>
          <p>{{ backupResult.message }}</p>
          <p v-if="backupResult.backupPath">备份文件: {{ backupResult.backupPath }}</p>
        </div>
      </div>

      <!-- 恢复标签页 -->
      <div v-if="activeTab === 'restore'" class="tab-pane">
        <h2>恢复备份</h2>
        <form @submit.prevent="restoreBackup">
          <div class="form-group">
            <label for="backupFile">备份文件:</label>
            <div class="input-with-button">
              <input 
                type="text" 
                id="backupFile" 
                v-model="restoreForm.backupFile"
                placeholder="请选择备份文件"
                required
              />
              <button type="button" class="btn select-btn" @click="openBackupFileDialog">
                选择
              </button>
            </div>
          </div>
          <div class="form-group">
            <label for="targetDir">目标目录:</label>
            <div class="input-with-button">
              <input 
                type="text" 
                id="targetDir" 
                v-model="restoreForm.targetDir"
                placeholder="请选择恢复目标路径"
                required
              />
              <button type="button" class="btn select-btn" @click="openTargetDirDialog">
                选择
              </button>
            </div>
          </div>
          <button type="submit" class="btn" :disabled="isLoading">
            {{ isLoading ? '恢复中...' : '开始恢复' }}
          </button>
        </form>
        <div v-if="restoreResult" class="result">
          <h3>恢复结果:</h3>
          <p>{{ restoreResult.message }}</p>
          <p v-if="restoreResult.targetDir">恢复到: {{ restoreResult.targetDir }}</p>
        </div>
      </div>

      <!-- 计划任务标签页 -->
      <div v-if="activeTab === 'schedule'" class="tab-pane">
        <h2>计划备份任务</h2>
        <button type="button" class="btn add-schedule-btn" @click="showScheduleDialog = true">
          创建计划任务
        </button>

        <h3>已创建的计划任务</h3>
        <div v-if="schedules.length === 0" class="no-data">
          暂无计划任务
        </div>
        <div v-else class="schedule-list">
          <div 
            v-for="schedule in schedules" 
            :key="schedule.id"
            class="schedule-item"
          >
            <div class="schedule-info">
              <p><strong>任务名称:</strong> {{ schedule.taskName }}</p>
              <p><strong>源目录:</strong> {{ schedule.sourceDir }}</p>
              <p><strong>输出目录:</strong> {{ schedule.outputDir }}</p>
              <p><strong>Cron表达式:</strong> {{ schedule.cronExpr }}</p>
              <p><strong>保留份数:</strong> {{ schedule.keepCopies || 0 }}</p>
              <p><strong>OSS配置:</strong> {{ getOSSConfigName(schedule.ossConfigId) || '未关联' }}</p>
              <p><strong>创建时间:</strong> {{ formatDate(schedule.createdAt) }}</p>
            </div>
            <div class="schedule-actions">
              <button 
                class="btn trigger-btn" 
                @click="triggerSchedule(schedule.id)"
                :disabled="isLoading"
              >
                触发
              </button>
              <button 
                class="btn edit-btn" 
                @click="editSchedule(schedule)"
                :disabled="isLoading"
              >
                编辑
              </button>
              <button 
                class="btn delete-btn" 
                @click="deleteSchedule(schedule.id)"
                :disabled="isLoading"
              >
                删除
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 备份记录标签页 -->
      <div v-if="activeTab === 'backup-records'" class="tab-pane">
      <h2>备份记录</h2>
      
      <!-- 筛选控件 -->
      <div class="filter-controls">
        <div class="form-group">
          <label for="scheduleFilter">按计划任务筛选:</label>
          <select 
            id="scheduleFilter" 
            v-model="backupFilter.scheduleId"
            @change="onFilterChange"
          >
            <option value="">全部</option>
            <option 
              v-for="schedule in schedules" 
              :key="schedule.id"
              :value="schedule.id"
            >
              {{ schedule.taskName }}
            </option>
          </select>
        </div>
      </div>
      
      <div v-if="backupHistory.length === 0" class="no-data">
        暂无备份记录
      </div>
      <div v-else class="backup-history-list">
            <div 
              v-for="record in backupHistory" 
              :key="record.id"
              class="backup-record-item"
            >
              <div class="backup-record-info">
                <p><strong>文件名:</strong> {{ record.fileName }}</p>
                <p><strong>源目录:</strong> {{ record.sourceDir }}</p>
                <p><strong>输出目录:</strong> {{ record.outputDir }}</p>
                <p><strong>创建时间:</strong> {{ formatDate(record.createdAt) }}</p>
                <p><strong>任务:</strong> {{ getScheduleName(record.scheduleId) || '手动备份' }}</p>
              </div>
              <div class="backup-record-actions">
                <button class="btn restore-btn" @click="restoreFromBackup(record.fileName)">恢复</button>
                <button class="btn download-btn" @click="downloadBackup(record.fileName)">下载</button>
                <button class="btn delete-btn" @click="deleteBackup(record.fileName)">删除</button>
              </div>
            </div>
          </div>
      
      <!-- 分页控件 -->
      <div v-if="backupFilter.total > 0" class="pagination">
        <button 
          class="btn" 
          @click="changePage(1)"
          :disabled="backupFilter.page === 1"
        >
          首页
        </button>
        <button 
          class="btn" 
          @click="changePage(backupFilter.page - 1)"
          :disabled="backupFilter.page === 1"
        >
          上一页
        </button>
        <span class="page-info">
          第 {{ backupFilter.page }} / {{ backupFilter.totalPages }} 页
        </span>
        <button 
          class="btn" 
          @click="changePage(backupFilter.page + 1)"
          :disabled="backupFilter.page === backupFilter.totalPages"
        >
          下一页
        </button>
        <button 
          class="btn" 
          @click="changePage(backupFilter.totalPages)"
          :disabled="backupFilter.page === backupFilter.totalPages"
        >
          末页
        </button>
        <div class="page-size">
          <label for="pageSize">每页条数:</label>
          <select 
            id="pageSize" 
            v-model.number="backupFilter.pageSize"
            @change="onPageSizeChange"
          >
            <option value="5">5</option>
            <option value="10">10</option>
            <option value="20">20</option>
            <option value="50">50</option>
          </select>
        </div>
      </div>
    </div>

      <!-- OSS配置标签页 -->
      <div v-if="activeTab === 'oss-config'" class="tab-pane">
        <h2>OSS配置</h2>
        <button class="btn add-schedule-btn" @click="openOSSConfigDialog">添加OSS配置</button>
        
        <div v-if="ossConfigs.length === 0" class="no-data">
          暂无OSS配置
        </div>
        <div v-else class="oss-config-list">
          <div v-for="config in ossConfigs" :key="config.id" class="oss-config-card">
            <div class="oss-config-info">
              <h3>{{ config.name }}</h3>
              <div class="config-detail">
                <p><strong>Endpoint:</strong> {{ config.endpoint }}</p>
                <p><strong>Access Key ID:</strong> {{ config.accessKeyId }}</p>
                <p><strong>Bucket名称:</strong> {{ config.bucketName }}</p>
                <p><strong>前缀:</strong> {{ config.prefix || '-' }}</p>
              </div>
            </div>
            <div class="oss-config-actions">
              <button class="btn" @click="editOSSConfig(config)">编辑</button>
              <button class="btn" @click="testExistingOSSConfig(config)">测试连接</button>
              <button class="btn delete-btn" @click="deleteOSSConfig(config.id)">删除</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 目录选择弹窗 -->
    <div v-if="showSourceDirDialog" class="modal">
      <div class="modal-content directory-modal">
        <h3>选择源目录</h3>
        <div class="dialog-body directory-selector">
          <div class="directory-tree">
            <div v-for="item in directoryTree" :key="item.path" class="directory-item">
              <div 
                class="directory-name" 
                @click="item.isDir ? (() => { toggleDirectory(item); this.selectedDirectory = item; })() : null"
                :class="{ selected: selectedDirectory && selectedDirectory.path === item.path }"
              >
                <span v-if="item.isDir" class="dir-icon">{{ item.expanded ? '▼' : '►' }}</span>
                <span v-else class="file-icon">📄</span>
                {{ item.name }}
              </div>
              <div v-if="item.isDir && item.expanded" class="subdirectories">
                <div v-for="child in item.children" :key="child.path" class="directory-item">
                  <div 
                    class="directory-name" 
                    @click="child.isDir ? (() => { toggleDirectory(child); this.selectedDirectory = child; })() : null"
                    :class="{ selected: selectedDirectory && selectedDirectory.path === child.path }"
                  >
                    <span v-if="child.isDir" class="dir-icon">{{ child.expanded ? '▼' : '►' }}</span>
                    <span v-else class="file-icon">📄</span>
                    {{ child.name }}
                  </div>
                  <div v-if="child.isDir && child.expanded" class="subdirectories">
                    <div v-for="grandchild in child.children" :key="grandchild.path" class="directory-item">
                      <div 
                        class="directory-name" 
                        @click="grandchild.isDir ? (() => { toggleDirectory(grandchild); this.selectedDirectory = grandchild; })() : null"
                        :class="{ selected: selectedDirectory && selectedDirectory.path === grandchild.path }"
                      >
                        <span v-if="grandchild.isDir" class="dir-icon">{{ grandchild.expanded ? '▼' : '►' }}</span>
                        <span v-else class="file-icon">📄</span>
                        {{ grandchild.name }}
                      </div>
                      <div v-if="grandchild.isDir && grandchild.expanded" class="subdirectories">
                        <div v-for="greatgrandchild in grandchild.children" :key="greatgrandchild.path" class="directory-item">
                          <div 
                            class="directory-name" 
                            @click="greatgrandchild.isDir ? (() => { toggleDirectory(greatgrandchild); this.selectedDirectory = greatgrandchild; })() : null"
                            :class="{ selected: selectedDirectory && selectedDirectory.path === greatgrandchild.path }"
                          >
                            <span v-if="greatgrandchild.isDir" class="dir-icon">{{ greatgrandchild.expanded ? '▼' : '►' }}</span>
                            <span v-else class="file-icon">📄</span>
                            {{ greatgrandchild.name }}
                          </div>
                          <div v-if="greatgrandchild.isDir && greatgrandchild.expanded" class="subdirectories">
                            <div v-for="greatgreatgrandchild in greatgrandchild.children" :key="greatgreatgrandchild.path" class="directory-item">
                              <div 
                                class="directory-name" 
                                @click="greatgreatgrandchild.isDir ? (() => { toggleDirectory(greatgreatgrandchild); this.selectedDirectory = greatgreatgrandchild; })() : null"
                                :class="{ selected: selectedDirectory && selectedDirectory.path === greatgreatgrandchild.path }"
                              >
                                <span v-if="greatgreatgrandchild.isDir" class="dir-icon">{{ greatgreatgrandchild.expanded ? '▼' : '►' }}</span>
                                <span v-else class="file-icon">📄</span>
                                {{ greatgreatgrandchild.name }}
                              </div>
                              <div v-if="greatgreatgrandchild.isDir && greatgreatgrandchild.expanded" class="subdirectories">
                                <div v-for="greatgreatgreatgrandchild in greatgreatgrandchild.children" :key="greatgreatgreatgrandchild.path" class="directory-item">
                                  <div 
                                    class="directory-name" 
                                    @click="greatgreatgreatgrandchild.isDir ? (() => { toggleDirectory(greatgreatgreatgrandchild); this.selectedDirectory = greatgreatgreatgrandchild; })() : null"
                                    :class="{ selected: selectedDirectory && selectedDirectory.path === greatgreatgreatgrandchild.path }"
                                  >
                                    <span v-if="greatgreatgreatgrandchild.isDir" class="dir-icon">{{ greatgreatgreatgrandchild.expanded ? '▼' : '►' }}</span>
                                    <span v-else class="file-icon">📄</span>
                                    {{ greatgreatgreatgrandchild.name }}
                                  </div>
                                  <div v-if="greatgreatgreatgrandchild.isDir && greatgreatgreatgrandchild.expanded" class="subdirectories">
                                    <div v-for="deepchild in greatgreatgreatgrandchild.children" :key="deepchild.path" class="directory-item">
                                      <div 
                                        class="directory-name" 
                                        @click="deepchild.isDir ? (() => { toggleDirectory(deepchild); this.selectedDirectory = deepchild; })() : null"
                                        :class="{ selected: selectedDirectory && selectedDirectory.path === deepchild.path }"
                                      >
                                        <span v-if="deepchild.isDir" class="dir-icon">{{ deepchild.expanded ? '▼' : '►' }}</span>
                                        <span v-else class="file-icon">📄</span>
                                        {{ deepchild.name }}
                                      </div>
                                      <div v-if="deepchild.isDir && deepchild.expanded" class="subdirectories">
                                        <!-- 可以继续添加更深层次的目录渲染 -->
                                      </div>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showSourceDirDialog = false">取消</button>
          <button class="btn" @click="selectSourceDir" :disabled="!selectedDirectory">
            确定
          </button>
        </div>
      </div>
    </div>

    <!-- 备份文件选择弹窗 -->
    <div v-if="showBackupFileDialog" class="modal">
      <div class="modal-content directory-modal">
        <h3>选择备份文件</h3>
        <div class="dialog-body directory-selector">
          <div class="directory-tree">
            <div v-for="item in directoryTree" :key="item.path" class="directory-item">
              <div 
                class="directory-name" 
                @click="item.isDir ? (() => { toggleDirectory(item); this.selectedDirectory = item; })() : selectFile(item)"
                :class="{ selected: (selectedDirectory && selectedDirectory.path === item.path) || (selectedFile && selectedFile.path === item.path) }"
              >
                <span v-if="item.isDir" class="dir-icon">{{ item.expanded ? '▼' : '►' }}</span>
                <span v-else class="file-icon">📄</span>
                {{ item.name }}
              </div>
              <div v-if="item.isDir && item.expanded" class="subdirectories">
                <div v-for="child in item.children" :key="child.path" class="directory-item">
                  <div 
                    class="directory-name" 
                    @click="child.isDir ? (() => { toggleDirectory(child); this.selectedDirectory = child; })() : selectFile(child)"
                    :class="{ selected: (selectedDirectory && selectedDirectory.path === child.path) || (selectedFile && selectedFile.path === child.path) }"
                  >
                    <span v-if="child.isDir" class="dir-icon">{{ child.expanded ? '▼' : '►' }}</span>
                    <span v-else class="file-icon">📄</span>
                    {{ child.name }}
                  </div>
                  <div v-if="child.isDir && child.expanded" class="subdirectories">
                    <div v-for="grandchild in child.children" :key="grandchild.path" class="directory-item">
                      <div 
                        class="directory-name" 
                        @click="grandchild.isDir ? (() => { toggleDirectory(grandchild); this.selectedDirectory = grandchild; })() : selectFile(grandchild)"
                        :class="{ selected: (selectedDirectory && selectedDirectory.path === grandchild.path) || (selectedFile && selectedFile.path === grandchild.path) }"
                      >
                        <span v-if="grandchild.isDir" class="dir-icon">{{ grandchild.expanded ? '▼' : '►' }}</span>
                        <span v-else class="file-icon">📄</span>
                        {{ grandchild.name }}
                      </div>
                      <div v-if="grandchild.isDir && grandchild.expanded" class="subdirectories">
                        <div v-for="greatgrandchild in grandchild.children" :key="greatgrandchild.path" class="directory-item">
                          <div 
                            class="directory-name" 
                            @click="greatgrandchild.isDir ? (() => { toggleDirectory(greatgrandchild); this.selectedDirectory = greatgrandchild; })() : selectFile(greatgrandchild)"
                            :class="{ selected: (selectedDirectory && selectedDirectory.path === greatgrandchild.path) || (selectedFile && selectedFile.path === greatgrandchild.path) }"
                          >
                            <span v-if="greatgrandchild.isDir" class="dir-icon">{{ greatgrandchild.expanded ? '▼' : '►' }}</span>
                            <span v-else class="file-icon">📄</span>
                            {{ greatgrandchild.name }}
                          </div>
                          <div v-if="greatgrandchild.isDir && greatgrandchild.expanded" class="subdirectories">
                            <div v-for="greatgreatgrandchild in greatgrandchild.children" :key="greatgreatgrandchild.path" class="directory-item">
                              <div 
                                class="directory-name" 
                                @click="greatgreatgrandchild.isDir ? (() => { toggleDirectory(greatgreatgrandchild); this.selectedDirectory = greatgreatgrandchild; })() : selectFile(greatgreatgrandchild)"
                                :class="{ selected: (selectedDirectory && selectedDirectory.path === greatgreatgrandchild.path) || (selectedFile && selectedFile.path === greatgreatgrandchild.path) }"
                              >
                                <span v-if="greatgreatgrandchild.isDir" class="dir-icon">{{ greatgreatgrandchild.expanded ? '▼' : '►' }}</span>
                                <span v-else class="file-icon">📄</span>
                                {{ greatgreatgrandchild.name }}
                              </div>
                              <div v-if="greatgreatgrandchild.isDir && greatgreatgrandchild.expanded" class="subdirectories">
                                <div v-for="greatgreatgreatgrandchild in greatgreatgrandchild.children" :key="greatgreatgreatgrandchild.path" class="directory-item">
                                  <div 
                                    class="directory-name" 
                                    @click="greatgreatgreatgrandchild.isDir ? (() => { toggleDirectory(greatgreatgreatgrandchild); this.selectedDirectory = greatgreatgreatgrandchild; })() : selectFile(greatgreatgreatgrandchild)"
                                    :class="{ selected: (selectedDirectory && selectedDirectory.path === greatgreatgreatgrandchild.path) || (selectedFile && selectedFile.path === greatgreatgreatgrandchild.path) }"
                                  >
                                    <span v-if="greatgreatgreatgrandchild.isDir" class="dir-icon">{{ greatgreatgreatgrandchild.expanded ? '▼' : '►' }}</span>
                                    <span v-else class="file-icon">📄</span>
                                    {{ greatgreatgreatgrandchild.name }}
                                  </div>
                                  <div v-if="greatgreatgreatgrandchild.isDir && greatgreatgreatgrandchild.expanded" class="subdirectories">
                                    <div v-for="deepchild in greatgreatgreatgrandchild.children" :key="deepchild.path" class="directory-item">
                                      <div 
                                        class="directory-name" 
                                        @click="deepchild.isDir ? (() => { toggleDirectory(deepchild); this.selectedDirectory = deepchild; })() : selectFile(deepchild)"
                                        :class="{ selected: (selectedDirectory && selectedDirectory.path === deepchild.path) || (selectedFile && selectedFile.path === deepchild.path) }"
                                      >
                                        <span v-if="deepchild.isDir" class="dir-icon">{{ deepchild.expanded ? '▼' : '►' }}</span>
                                        <span v-else class="file-icon">📄</span>
                                        {{ deepchild.name }}
                                      </div>
                                      <div v-if="deepchild.isDir && deepchild.expanded" class="subdirectories">
                                        <!-- 可以继续添加更深层次的目录渲染 -->
                                      </div>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showBackupFileDialog = false">取消</button>
          <button class="btn" @click="selectBackupFile" :disabled="!selectedFile">
            确定
          </button>
        </div>
      </div>
    </div>

    <!-- 目标目录选择弹窗 -->
    <div v-if="showTargetDirDialog" class="modal">
      <div class="modal-content directory-modal">
        <h3>选择目标目录</h3>
        <div class="dialog-body directory-selector">
          <div class="directory-tree">
            <div v-for="item in directoryTree" :key="item.path" class="directory-item">
              <div 
                class="directory-name" 
                @click="item.isDir ? (() => { toggleDirectory(item); this.selectedDirectory = item; })() : null"
                :class="{ selected: selectedDirectory && selectedDirectory.path === item.path }"
              >
                <span v-if="item.isDir" class="dir-icon">{{ item.expanded ? '▼' : '►' }}</span>
                <span v-else class="file-icon">📄</span>
                {{ item.name }}
              </div>
              <div v-if="item.isDir && item.expanded" class="subdirectories">
                <div v-for="child in item.children" :key="child.path" class="directory-item">
                  <div 
                    class="directory-name" 
                    @click="child.isDir ? (() => { toggleDirectory(child); this.selectedDirectory = child; })() : null"
                    :class="{ selected: selectedDirectory && selectedDirectory.path === child.path }"
                  >
                    <span v-if="child.isDir" class="dir-icon">{{ child.expanded ? '▼' : '►' }}</span>
                    <span v-else class="file-icon">📄</span>
                    {{ child.name }}
                  </div>
                  <div v-if="child.isDir && child.expanded" class="subdirectories">
                    <div v-for="grandchild in child.children" :key="grandchild.path" class="directory-item">
                      <div 
                        class="directory-name" 
                        @click="grandchild.isDir ? (() => { toggleDirectory(grandchild); this.selectedDirectory = grandchild; })() : null"
                        :class="{ selected: selectedDirectory && selectedDirectory.path === grandchild.path }"
                      >
                        <span v-if="grandchild.isDir" class="dir-icon">{{ grandchild.expanded ? '▼' : '►' }}</span>
                        <span v-else class="file-icon">📄</span>
                        {{ grandchild.name }}
                      </div>
                      <div v-if="grandchild.isDir && grandchild.expanded" class="subdirectories">
                        <div v-for="greatgrandchild in grandchild.children" :key="greatgrandchild.path" class="directory-item">
                          <div 
                            class="directory-name" 
                            @click="greatgrandchild.isDir ? (() => { toggleDirectory(greatgrandchild); this.selectedDirectory = greatgrandchild; })() : null"
                            :class="{ selected: selectedDirectory && selectedDirectory.path === greatgrandchild.path }"
                          >
                            <span v-if="greatgrandchild.isDir" class="dir-icon">{{ greatgrandchild.expanded ? '▼' : '►' }}</span>
                            <span v-else class="file-icon">📄</span>
                            {{ greatgrandchild.name }}
                          </div>
                          <div v-if="greatgrandchild.isDir && greatgrandchild.expanded" class="subdirectories">
                            <div v-for="greatgreatgrandchild in greatgrandchild.children" :key="greatgreatgrandchild.path" class="directory-item">
                              <div 
                                class="directory-name" 
                                @click="greatgreatgrandchild.isDir ? (() => { toggleDirectory(greatgreatgrandchild); this.selectedDirectory = greatgreatgrandchild; })() : null"
                                :class="{ selected: selectedDirectory && selectedDirectory.path === greatgreatgrandchild.path }"
                              >
                                <span v-if="greatgreatgrandchild.isDir" class="dir-icon">{{ greatgreatgrandchild.expanded ? '▼' : '►' }}</span>
                                <span v-else class="file-icon">📄</span>
                                {{ greatgreatgrandchild.name }}
                              </div>
                              <div v-if="greatgreatgrandchild.isDir && greatgreatgrandchild.expanded" class="subdirectories">
                                <div v-for="greatgreatgreatgrandchild in greatgreatgrandchild.children" :key="greatgreatgreatgrandchild.path" class="directory-item">
                                  <div 
                                    class="directory-name" 
                                    @click="greatgreatgreatgrandchild.isDir ? (() => { toggleDirectory(greatgreatgreatgrandchild); this.selectedDirectory = greatgreatgreatgrandchild; })() : null"
                                    :class="{ selected: selectedDirectory && selectedDirectory.path === greatgreatgreatgrandchild.path }"
                                  >
                                    <span v-if="greatgreatgreatgrandchild.isDir" class="dir-icon">{{ greatgreatgreatgrandchild.expanded ? '▼' : '►' }}</span>
                                    <span v-else class="file-icon">📄</span>
                                    {{ greatgreatgreatgrandchild.name }}
                                  </div>
                                  <div v-if="greatgreatgreatgrandchild.isDir && greatgreatgreatgrandchild.expanded" class="subdirectories">
                                    <div v-for="deepchild in greatgreatgreatgrandchild.children" :key="deepchild.path" class="directory-item">
                                      <div 
                                        class="directory-name" 
                                        @click="deepchild.isDir ? (() => { toggleDirectory(deepchild); this.selectedDirectory = deepchild; })() : null"
                                        :class="{ selected: selectedDirectory && selectedDirectory.path === deepchild.path }"
                                      >
                                        <span v-if="deepchild.isDir" class="dir-icon">{{ deepchild.expanded ? '▼' : '►' }}</span>
                                        <span v-else class="file-icon">📄</span>
                                        {{ deepchild.name }}
                                      </div>
                                      <div v-if="deepchild.isDir && deepchild.expanded" class="subdirectories">
                                        <!-- 可以继续添加更深层次的目录渲染 -->
                                      </div>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showTargetDirDialog = false">取消</button>
          <button class="btn" @click="selectTargetDir" :disabled="!selectedDirectory">
            确定
          </button>
        </div>
      </div>
    </div>

    <!-- 计划任务创建弹窗 -->
    <div v-if="showScheduleDialog" class="modal">
      <div class="modal-content">
        <h3>{{ editingScheduleId ? '编辑计划任务' : '创建计划任务' }}</h3>
        <div class="dialog-body">
          <div class="form-group">
            <label for="taskName">任务名称:</label>
            <input 
              type="text" 
              id="taskName" 
              v-model="scheduleForm.taskName"
              placeholder="请输入任务名称"
              required
            />
          </div>
          <div class="form-group">
            <label for="scheduleSourceDir">源目录:</label>
            <div class="input-with-button">
              <input 
                type="text" 
                id="scheduleSourceDir" 
                v-model="scheduleForm.sourceDir"
                placeholder="请选择要备份的目录"
                required
              />
              <button type="button" class="btn select-btn" @click="openScheduleSourceDirDialog">
                选择
              </button>
            </div>
          </div>
          <div class="form-group">
            <label for="scheduleOutputDir">输出目录:</label>
            <input 
              type="text" 
              id="scheduleOutputDir" 
              v-model="scheduleForm.outputDir"
              placeholder="备份文件保存路径"
              readonly
            />
          </div>
          <div class="form-group">
            <label for="cronExpr">Cron表达式:</label>
            <input 
              type="text" 
              id="cronExpr" 
              v-model="scheduleForm.cronExpr"
              placeholder="例如: 0 0 * * * (每天凌晨执行)"
              required
            />
          </div>
          <div class="form-group">
            <label for="keepCopies">保留份数:</label>
            <input 
              type="number" 
              id="keepCopies" 
              v-model.number="scheduleForm.keepCopies"
              placeholder="保留最近备份的数量"
              min="0"
            />
          </div>
          <div class="form-group">
            <label for="ossConfigId">OSS配置:</label>
            <select 
              id="ossConfigId" 
              v-model="scheduleForm.ossConfigId"
            >
              <option value="">不使用OSS</option>
              <option 
                v-for="config in ossConfigs" 
                :key="config.id"
                :value="config.id"
              >
                {{ config.name }}
              </option>
            </select>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showScheduleDialog = false">取消</button>
          <button class="btn" @click="createSchedule" :disabled="isLoading">
            {{ isLoading ? (editingScheduleId ? '保存中...' : '创建中...') : (editingScheduleId ? '保存' : '确定') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 计划任务源目录选择弹窗 -->
    <div v-if="showScheduleSourceDirDialog" class="modal">
      <div class="modal-content directory-modal">
        <h3>选择源目录</h3>
        <div class="dialog-body directory-selector">
          <div class="directory-tree">
            <div v-for="item in directoryTree" :key="item.path" class="directory-item">
              <div 
                class="directory-name" 
                @click="item.isDir ? (() => { toggleDirectory(item); this.selectedDirectory = item; })() : null"
                :class="{ selected: selectedDirectory && selectedDirectory.path === item.path }"
              >
                <span v-if="item.isDir" class="dir-icon">{{ item.expanded ? '▼' : '►' }}</span>
                <span v-else class="file-icon">📄</span>
                {{ item.name }}
              </div>
              <div v-if="item.isDir && item.expanded" class="subdirectories">
                <div v-for="child in item.children" :key="child.path" class="directory-item">
                  <div 
                    class="directory-name" 
                    @click="child.isDir ? (() => { toggleDirectory(child); this.selectedDirectory = child; })() : null"
                    :class="{ selected: selectedDirectory && selectedDirectory.path === child.path }"
                  >
                    <span v-if="child.isDir" class="dir-icon">{{ child.expanded ? '▼' : '►' }}</span>
                    <span v-else class="file-icon">📄</span>
                    {{ child.name }}
                  </div>
                  <div v-if="child.isDir && child.expanded" class="subdirectories">
                    <div v-for="grandchild in child.children" :key="grandchild.path" class="directory-item">
                      <div 
                        class="directory-name" 
                        @click="grandchild.isDir ? (() => { toggleDirectory(grandchild); this.selectedDirectory = grandchild; })() : null"
                        :class="{ selected: selectedDirectory && selectedDirectory.path === grandchild.path }"
                      >
                        <span v-if="grandchild.isDir" class="dir-icon">{{ grandchild.expanded ? '▼' : '►' }}</span>
                        <span v-else class="file-icon">📄</span>
                        {{ grandchild.name }}
                      </div>
                      <div v-if="grandchild.isDir && grandchild.expanded" class="subdirectories">
                        <div v-for="greatgrandchild in grandchild.children" :key="greatgrandchild.path" class="directory-item">
                          <div 
                            class="directory-name" 
                            @click="greatgrandchild.isDir ? (() => { toggleDirectory(greatgrandchild); this.selectedDirectory = greatgrandchild; })() : null"
                            :class="{ selected: selectedDirectory && selectedDirectory.path === greatgrandchild.path }"
                          >
                            <span v-if="greatgrandchild.isDir" class="dir-icon">{{ greatgrandchild.expanded ? '▼' : '►' }}</span>
                            <span v-else class="file-icon">📄</span>
                            {{ greatgrandchild.name }}
                          </div>
                          <div v-if="greatgrandchild.isDir && greatgrandchild.expanded" class="subdirectories">
                            <div v-for="greatgreatgrandchild in greatgrandchild.children" :key="greatgreatgrandchild.path" class="directory-item">
                              <div 
                                class="directory-name" 
                                @click="greatgreatgrandchild.isDir ? (() => { toggleDirectory(greatgreatgrandchild); this.selectedDirectory = greatgreatgrandchild; })() : null"
                                :class="{ selected: selectedDirectory && selectedDirectory.path === greatgreatgrandchild.path }"
                              >
                                <span v-if="greatgreatgrandchild.isDir" class="dir-icon">{{ greatgreatgrandchild.expanded ? '▼' : '►' }}</span>
                                <span v-else class="file-icon">📄</span>
                                {{ greatgreatgrandchild.name }}
                              </div>
                              <div v-if="greatgreatgrandchild.isDir && greatgreatgrandchild.expanded" class="subdirectories">
                                <div v-for="greatgreatgreatgrandchild in greatgreatgrandchild.children" :key="greatgreatgreatgrandchild.path" class="directory-item">
                                  <div 
                                    class="directory-name" 
                                    @click="greatgreatgreatgrandchild.isDir ? (() => { toggleDirectory(greatgreatgreatgrandchild); this.selectedDirectory = greatgreatgreatgrandchild; })() : null"
                                    :class="{ selected: selectedDirectory && selectedDirectory.path === greatgreatgreatgrandchild.path }"
                                  >
                                    <span v-if="greatgreatgreatgrandchild.isDir" class="dir-icon">{{ greatgreatgreatgrandchild.expanded ? '▼' : '►' }}</span>
                                    <span v-else class="file-icon">📄</span>
                                    {{ greatgreatgreatgrandchild.name }}
                                  </div>
                                  <div v-if="greatgreatgreatgrandchild.isDir && greatgreatgreatgrandchild.expanded" class="subdirectories">
                                    <div v-for="deepchild in greatgreatgreatgrandchild.children" :key="deepchild.path" class="directory-item">
                                      <div 
                                        class="directory-name" 
                                        @click="deepchild.isDir ? (() => { toggleDirectory(deepchild); this.selectedDirectory = deepchild; })() : null"
                                        :class="{ selected: selectedDirectory && selectedDirectory.path === deepchild.path }"
                                      >
                                        <span v-if="deepchild.isDir" class="dir-icon">{{ deepchild.expanded ? '▼' : '►' }}</span>
                                        <span v-else class="file-icon">📄</span>
                                        {{ deepchild.name }}
                                      </div>
                                      <div v-if="deepchild.isDir && deepchild.expanded" class="subdirectories">
                                        <!-- 可以继续添加更深层次的目录渲染 -->
                                      </div>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showScheduleSourceDirDialog = false">取消</button>
          <button class="btn" @click="selectScheduleSourceDir" :disabled="!selectedDirectory">
            确定
          </button>
        </div>
      </div>
    </div>

    <!-- OSS配置弹窗 -->
    <div v-if="showOSSConfigDialog" class="modal">
      <div class="modal-content">
        <h3>{{ editingOSSConfigId ? '编辑OSS配置' : '添加OSS配置' }}</h3>
        <div class="dialog-body">
          <div class="form-group">
            <label for="ossName">配置名称:</label>
            <input 
              type="text" 
              id="ossName" 
              v-model="ossConfigForm.name"
              placeholder="请输入配置名称"
              required
            />
          </div>
          <div class="form-group">
            <label for="ossEndpoint">Endpoint:</label>
            <input 
              type="text" 
              id="ossEndpoint" 
              v-model="ossConfigForm.endpoint"
              placeholder="例如: oss-cn-hangzhou.aliyuncs.com"
              required
            />
          </div>
          <div class="form-group">
            <label for="ossAccessKeyId">Access Key ID:</label>
            <input 
              type="text" 
              id="ossAccessKeyId" 
              v-model="ossConfigForm.accessKeyId"
              placeholder="请输入Access Key ID"
              required
            />
          </div>
          <div class="form-group">
            <label for="ossAccessKeySecret">Access Key Secret:</label>
            <input 
              type="text" 
              id="ossAccessKeySecret" 
              v-model="ossConfigForm.accessKeySecret"
              placeholder="请输入Access Key Secret"
              required
            />
          </div>
          <div class="form-group">
            <label for="ossBucketName">Bucket名称:</label>
            <input 
              type="text" 
              id="ossBucketName" 
              v-model="ossConfigForm.bucketName"
              placeholder="请输入Bucket名称"
              required
            />
          </div>
          <div class="form-group">
            <label for="ossPrefix">前缀 (可选):</label>
            <input 
              type="text" 
              id="ossPrefix" 
              v-model="ossConfigForm.prefix"
              placeholder="例如: backup/"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="showOSSConfigDialog = false">取消</button>
          <button class="btn" @click="testOSSConfig" :disabled="isLoading">
            {{ isLoading ? '测试中...' : '测试连接' }}
          </button>
          <button class="btn" @click="saveOSSConfig" :disabled="isLoading">
            {{ isLoading ? (editingOSSConfigId ? '保存中...' : '创建中...') : (editingOSSConfigId ? '保存' : '确定') }}
          </button>
        </div>
      </div>
    </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      isLoggedIn: false,
      loginPassword: '',
      loginError: '',
      token: '',
      activeTab: 'backup',
      tabs: [
        { id: 'backup', name: '备份' },
        { id: 'restore', name: '恢复' },
        { id: 'schedule', name: '计划任务' },
        { id: 'backup-records', name: '备份记录' },
        { id: 'oss-config', name: 'OSS配置' }
      ],
      backupForm: {
        sourceDir: '',
        outputDir: '',
        fileName: ''
      },
      restoreForm: {
        backupFile: '',
        targetDir: ''
      },
      scheduleForm: {
        sourceDir: '',
        outputDir: '',
        cronExpr: '',
        taskName: '',
        keepCopies: 5,
        ossConfigId: ''
      },
      isLoading: false,
      backupResult: null,
      restoreResult: null,
      schedules: [],
      // 弹窗相关
      showSourceDirDialog: false,
      showBackupFileDialog: false,
      showTargetDirDialog: false,
      showScheduleDialog: false,
      showScheduleSourceDirDialog: false,
      // 目录树相关
      directoryTree: [],
      selectedDirectory: null,
      selectedFile: null,
      // 编辑计划任务相关
      editingScheduleId: null,
      // 备份记录
      backupHistory: [],
      // 备份记录筛选和分页
      backupFilter: {
        scheduleId: '',
        page: 1,
        pageSize: 10,
        total: 0,
        totalPages: 0
      },
      // OSS配置相关
      ossConfigs: [],
      showOSSConfigDialog: false,
      ossConfigForm: {
        name: '',
        endpoint: '',
        accessKeyId: '',
        accessKeySecret: '',
        bucketName: '',
        prefix: ''
      },
      editingOSSConfigId: null
    }
  },
  mounted() {
    this.fetchSchedules()
    this.fetchBackupRecords()
    this.fetchOSSConfigs()
    // 设置默认输出目录为backup目录
    this.backupForm.outputDir = './backup'
    this.scheduleForm.outputDir = './backup'
  },
  watch: {
    activeTab(newTab) {
      if (newTab === 'backup-history') {
        this.fetchBackupRecords()
      }
    }
  },
  methods: {
    async fetchOSSConfigs() {
      try {
        const response = await fetch('/api/oss-configs', {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        const data = await response.json()
        if (response.ok) {
          this.ossConfigs = data.configs
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
        }
      } catch (error) {
        console.error('获取OSS配置失败:', error)
      }
    },
    async createBackup() {
      this.isLoading = true
      this.backupResult = null
      try {
        const response = await fetch('/api/backup', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
          },
          body: JSON.stringify(this.backupForm)
        })
        const data = await response.json()
        if (response.ok) {
          this.backupResult = data
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          this.backupResult = { message: '登录已过期，请重新登录' }
        } else {
          this.backupResult = { message: '备份失败: ' + (data.error || '未知错误') }
        }
      } catch (error) {
        this.backupResult = { message: '备份失败: ' + error.message }
      } finally {
        this.isLoading = false
      }
    },
    async restoreBackup() {
      this.isLoading = true
      this.restoreResult = null
      try {
        const response = await fetch('/api/restore', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
          },
          body: JSON.stringify(this.restoreForm)
        })
        const data = await response.json()
        if (response.ok) {
          this.restoreResult = data
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          this.restoreResult = { message: '登录已过期，请重新登录' }
        } else {
          this.restoreResult = { message: '恢复失败: ' + (data.error || '未知错误') }
        }
      } catch (error) {
        this.restoreResult = { message: '恢复失败: ' + error.message }
      } finally {
        this.isLoading = false
      }
    },
    async createSchedule() {
      this.isLoading = true
      try {
        const response = await fetch('/api/schedule', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
          },
          body: JSON.stringify(this.scheduleForm)
        })
        const data = await response.json()
        if (response.ok) {
          // 重置表单
          this.scheduleForm = {
            sourceDir: '',
            outputDir: './backup',
            cronExpr: '',
            taskName: ''
          }
          // 关闭弹窗
          this.showScheduleDialog = false
          // 重新获取计划任务列表
          this.fetchSchedules()
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          alert('创建计划任务失败: ' + (data.error || '未知错误'))
        }
      } catch (error) {
        alert('创建计划任务失败: ' + error.message)
      } finally {
        this.isLoading = false
      }
    },
    async fetchSchedules() {
      try {
        const response = await fetch('/api/schedules', {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        const data = await response.json()
        if (response.ok) {
          this.schedules = data.schedules
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
        }
      } catch (error) {
        console.error('获取计划任务失败:', error)
      }
    },
    async fetchBackupRecords() {
      try {
        // 构建查询参数
        const params = new URLSearchParams()
        if (this.backupFilter.scheduleId) {
          params.append('scheduleId', this.backupFilter.scheduleId)
        }
        params.append('page', this.backupFilter.page)
        params.append('pageSize', this.backupFilter.pageSize)

        const response = await fetch(`/api/backup-records?${params.toString()}`, {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        const data = await response.json()
        if (response.ok) {
          this.backupHistory = data.records
          this.backupFilter.total = data.total
          this.backupFilter.totalPages = data.totalPages
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
        }
      } catch (error) {
        console.error('获取备份记录失败:', error)
      }
    },
    async deleteBackup(fileName) {
      if (!confirm('确定要删除这个备份文件吗？')) {
        return
      }
      
      try {
        const response = await fetch(`/api/backup/${encodeURIComponent(fileName)}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        const data = await response.json()
        if (response.ok) {
          alert('删除成功')
          this.fetchBackupRecords()
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          alert('删除失败: ' + data.error)
        }
      } catch (error) {
        console.error('删除备份失败:', error)
        alert('删除备份失败: ' + error.message)
      }
    },
    restoreFromBackup(fileName) {
      // 打开恢复弹窗
      this.restoreForm.backupFile = fileName
      this.showRestoreDialog = true
    },
    async downloadBackup(fileName) {
      // 下载备份文件
      try {
        const response = await fetch(`/api/backup/${encodeURIComponent(fileName)}`, {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        if (response.ok) {
          // 处理文件下载
          const blob = await response.blob()
          const url = window.URL.createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url
          a.download = fileName
          document.body.appendChild(a)
          a.click()
          window.URL.revokeObjectURL(url)
          document.body.removeChild(a)
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          alert('下载失败')
        }
      } catch (error) {
        console.error('下载备份失败:', error)
        alert('下载备份失败: ' + error.message)
      }
    },
    onFilterChange() {
      // 筛选条件变化时，重置页码并重新获取数据
      this.backupFilter.page = 1
      this.fetchBackupRecords()
    },
    changePage(page) {
      // 改变页码并重新获取数据
      if (page >= 1 && page <= this.backupFilter.totalPages) {
        this.backupFilter.page = page
        this.fetchBackupRecords()
      }
    },
    onPageSizeChange() {
      // 每页条数变化时，重置页码并重新获取数据
      this.backupFilter.page = 1
      this.fetchBackupRecords()
    },
    // OSS配置相关方法
    openOSSConfigDialog() {
      // 重置表单
      this.ossConfigForm = {
        name: '',
        endpoint: '',
        accessKeyId: '',
        accessKeySecret: '',
        bucketName: '',
        prefix: ''
      }
      this.editingOSSConfigId = null
      this.showOSSConfigDialog = true
    },
    editOSSConfig(config) {
      // 填充表单
      this.ossConfigForm = {
        name: config.name,
        endpoint: config.endpoint,
        accessKeyId: config.accessKeyId,
        accessKeySecret: config.accessKeySecret,
        bucketName: config.bucketName,
        prefix: config.prefix
      }
      this.editingOSSConfigId = config.id
      this.showOSSConfigDialog = true
    },
    async saveOSSConfig() {
      // 表单验证
      if (!this.ossConfigForm.name) {
        alert('请输入配置名称')
        return
      }
      if (!this.ossConfigForm.endpoint) {
        alert('请输入OSS endpoint')
        return
      }
      if (!this.ossConfigForm.accessKeyId) {
        alert('请输入Access Key ID')
        return
      }
      if (!this.ossConfigForm.accessKeySecret) {
        alert('请输入Access Key Secret')
        return
      }
      if (!this.ossConfigForm.bucketName) {
        alert('请输入Bucket名称')
        return
      }
      
      try {
        let response
        if (this.editingOSSConfigId) {
          // 更新OSS配置
          response = await fetch(`/api/oss-config/${this.editingOSSConfigId}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${this.token}`
            },
            body: JSON.stringify(this.ossConfigForm)
          })
        } else {
          // 创建OSS配置
          response = await fetch('/api/oss-config', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${this.token}`
            },
            body: JSON.stringify(this.ossConfigForm)
          })
        }
        const data = await response.json()
        if (response.ok) {
          alert(data.message)
          this.showOSSConfigDialog = false
          this.fetchOSSConfigs()
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          alert('操作失败: ' + data.error)
        }
      } catch (error) {
        console.error('保存OSS配置失败:', error)
        alert('保存OSS配置失败: ' + error.message)
      }
    },
    async deleteOSSConfig(id) {
      if (!confirm('确定要删除这个OSS配置吗？')) {
        return
      }
      
      try {
        const response = await fetch(`/api/oss-config/${id}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        const data = await response.json()
        if (response.ok) {
          alert('删除成功')
          this.fetchOSSConfigs()
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          alert('删除失败: ' + data.error)
        }
      } catch (error) {
        console.error('删除OSS配置失败:', error)
        alert('删除OSS配置失败: ' + error.message)
      }
    },
    async testOSSConfig() {
      // 表单验证
      if (!this.ossConfigForm.name) {
        alert('请输入配置名称')
        return
      }
      if (!this.ossConfigForm.endpoint) {
        alert('请输入OSS endpoint')
        return
      }
      if (!this.ossConfigForm.accessKeyId) {
        alert('请输入Access Key ID')
        return
      }
      if (!this.ossConfigForm.accessKeySecret) {
        alert('请输入Access Key Secret')
        return
      }
      if (!this.ossConfigForm.bucketName) {
        alert('请输入Bucket名称')
        return
      }
      
      try {
        this.isLoading = true
        const response = await fetch('/api/oss-config/test', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
          },
          body: JSON.stringify(this.ossConfigForm)
        })
        const data = await response.json()
        if (response.ok) {
          alert('测试成功: ' + data.message)
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          alert('测试失败: ' + data.error)
        }
      } catch (error) {
        console.error('测试OSS配置失败:', error)
        alert('测试OSS配置失败: ' + error.message)
      } finally {
        this.isLoading = false
      }
    },
    async testExistingOSSConfig(config) {
      try {
        this.isLoading = true
        const response = await fetch('/api/oss-config/test', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
          },
          body: JSON.stringify({
            name: config.name,
            endpoint: config.endpoint,
            accessKeyId: config.accessKeyId,
            accessKeySecret: config.accessKeySecret,
            bucketName: config.bucketName,
            prefix: config.prefix
          })
        })
        const data = await response.json()
        if (response.ok) {
          alert('测试成功: ' + data.message)
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          alert('测试失败: ' + data.error)
        }
      } catch (error) {
        console.error('测试OSS配置失败:', error)
        alert('测试OSS配置失败: ' + error.message)
      } finally {
        this.isLoading = false
      }
    },
    getOSSConfigName(ossConfigId) {
      if (!ossConfigId) return ''
      const config = this.ossConfigs.find(config => config.id === ossConfigId)
      return config ? config.name : '未知配置'
    },
    getScheduleName(scheduleId) {
      if (!scheduleId) return ''
      const schedule = this.schedules.find(schedule => schedule.id === scheduleId)
      return schedule ? `${schedule.taskName} (${schedule.id})` : `未知任务 (${scheduleId})`
    },
    async login() {
      // 通过后端API验证密码
      try {
        this.isLoading = true
        const response = await fetch('/api/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ password: this.loginPassword })
        })
        const data = await response.json()
        if (response.ok) {
          this.isLoggedIn = true
          this.loginError = ''
          // 保存token
          this.token = 'admin123'
          // 可以将登录状态和token保存到localStorage，实现记住登录状态
          localStorage.setItem('isLoggedIn', 'true')
          localStorage.setItem('token', this.token)
        } else {
          this.loginError = data.error || '密码错误，请重新输入'
        }
      } catch (error) {
        console.error('登录失败:', error)
        this.loginError = '登录失败，请稍后重试'
      } finally {
        this.isLoading = false
      }
    },
    logout() {
      this.isLoggedIn = false
      this.loginPassword = ''
      this.token = ''
      // 清除localStorage中的登录状态和token
      localStorage.removeItem('isLoggedIn')
      localStorage.removeItem('token')
    },
    async deleteSchedule(id) {
      if (!confirm('确定要删除这个计划任务吗？')) {
        return
      }
      this.isLoading = true
      try {
        const response = await fetch(`/api/schedule/${id}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        if (response.ok) {
          // 重新获取计划任务列表
          this.fetchSchedules()
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          const data = await response.json()
          alert('删除计划任务失败: ' + (data.error || '未知错误'))
        }
      } catch (error) {
        alert('删除计划任务失败: ' + error.message)
      } finally {
        this.isLoading = false
      }
    },
    formatDate(dateString) {
      const date = new Date(dateString)
      return date.toLocaleString()
    },
    // 目录树相关方法
    async loadDirectories(path = '') {
      try {
        const response = await fetch(`/api/directories?path=${encodeURIComponent(path)}`, {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        })
        const data = await response.json()
        if (response.ok) {
          // 对目录进行排序：文件夹在前，然后按名称升序
          const sortedDirectories = data.directories.sort((a, b) => {
            // 先按是否为目录排序（目录在前）
            if (a.isDir && !b.isDir) return -1
            if (!a.isDir && b.isDir) return 1
            // 再按名称升序排序
            return a.name.localeCompare(b.name)
          })
          // 为每个目录添加expanded属性
          this.directoryTree = sortedDirectories.map(item => ({
            ...item,
            expanded: false,
            children: []
          }))
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
        }
      } catch (error) {
        console.error('加载目录失败:', error)
      }
    },
    async toggleDirectory(item) {
      if (!item.isDir) return
      
      item.expanded = !item.expanded
      if (item.expanded) {
        // 加载子目录
        try {
          const response = await fetch(`/api/directories?path=${encodeURIComponent(item.path)}`, {
            headers: {
              'Authorization': `Bearer ${this.token}`
            }
          })
          const data = await response.json()
          if (response.ok) {
            // 对目录进行排序：文件夹在前，然后按名称升序
            const sortedDirectories = data.directories.sort((a, b) => {
              // 先按是否为目录排序（目录在前）
              if (a.isDir && !b.isDir) return -1
              if (!a.isDir && b.isDir) return 1
              // 再按名称升序排序
              return a.name.localeCompare(b.name)
            })
            item.children = sortedDirectories.map(child => ({
              ...child,
              expanded: false,
              children: []
            }))
          } else if (response.status === 401) {
            // 未授权，跳转到登录页面
            this.isLoggedIn = false
            this.token = ''
            localStorage.removeItem('isLoggedIn')
            localStorage.removeItem('token')
          }
        } catch (error) {
          console.error('加载子目录失败:', error)
        }
      }
    },
    selectFile(item) {
      this.selectedFile = item
    },
    // 弹窗显示时加载目录
    openSourceDirDialog() {
      this.showSourceDirDialog = true
      this.selectedDirectory = null
      this.loadDirectories()
    },
    openBackupFileDialog() {
      this.showBackupFileDialog = true
      this.selectedFile = null
      this.loadDirectories()
    },
    openTargetDirDialog() {
      this.showTargetDirDialog = true
      this.selectedDirectory = null
      this.loadDirectories()
    },
    openScheduleSourceDirDialog() {
      this.showScheduleSourceDirDialog = true
      this.selectedDirectory = null
      this.loadDirectories()
    },
    // 目录选择方法
    selectSourceDir() {
      if (this.selectedDirectory) {
        this.backupForm.sourceDir = this.selectedDirectory.path
        this.showSourceDirDialog = false
        this.selectedDirectory = null
      }
    },
    selectBackupFile() {
      if (this.selectedFile) {
        this.restoreForm.backupFile = this.selectedFile.path
        this.showBackupFileDialog = false
        this.selectedFile = null
      }
    },
    selectTargetDir() {
      if (this.selectedDirectory) {
        this.restoreForm.targetDir = this.selectedDirectory.path
        this.showTargetDirDialog = false
        this.selectedDirectory = null
      }
    },
    selectScheduleSourceDir() {
      if (this.selectedDirectory) {
        this.scheduleForm.sourceDir = this.selectedDirectory.path
        this.showScheduleSourceDirDialog = false
        this.selectedDirectory = null
      }
    },
    async triggerSchedule(id) {
      this.isLoading = true
      try {
        const response = await fetch(`/api/schedule/${id}/trigger`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`
          },
          body: JSON.stringify({})
        })
        const data = await response.json()
        if (response.ok) {
          alert('计划任务已触发')
        } else if (response.status === 401) {
          // 未授权，跳转到登录页面
          this.isLoggedIn = false
          this.token = ''
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('token')
          alert('登录已过期，请重新登录')
        } else {
          alert('触发计划任务失败: ' + (data.error || '未知错误'))
        }
      } catch (error) {
        alert('触发计划任务失败: ' + error.message)
      } finally {
        this.isLoading = false
      }
    },
    editSchedule(schedule) {
      // 填充表单
      this.scheduleForm = {
        sourceDir: schedule.sourceDir,
        outputDir: schedule.outputDir,
        cronExpr: schedule.cronExpr,
        taskName: schedule.taskName,
        keepCopies: schedule.keepCopies || 5,
        ossConfigId: schedule.ossConfigId || ''
      }
      // 存储正在编辑的计划任务ID
      this.editingScheduleId = schedule.id
      // 打开弹窗
      this.showScheduleDialog = true
    },
    async createSchedule() {
      // 表单验证
      if (!this.scheduleForm.sourceDir) {
        alert('请选择源目录')
        return
      }
      if (!this.scheduleForm.cronExpr) {
        alert('请输入 Cron 表达式')
        return
      }
      if (!this.scheduleForm.taskName) {
        alert('请输入任务名称')
        return
      }
      
      this.isLoading = true
      try {
        let response
        if (this.editingScheduleId) {
          // 更新计划任务
          response = await fetch(`/api/schedule/${this.editingScheduleId}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${this.token}`
            },
            body: JSON.stringify(this.scheduleForm)
          })
        } else {
          // 创建计划任务
          response = await fetch('/api/schedule', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${this.token}`
            },
            body: JSON.stringify(this.scheduleForm)
          })
        }
        const data = await response.json()
        if (response.ok) {
          // 重置表单
          this.scheduleForm = {
            sourceDir: '',
            outputDir: './backup',
            cronExpr: '',
            taskName: ''
          }
          // 重置编辑状态
          this.editingScheduleId = null
          // 关闭弹窗
          this.showScheduleDialog = false
          // 重新获取计划任务列表
          this.fetchSchedules()
        } else {
          alert('操作失败: ' + (data.error || '未知错误'))
        }
      } catch (error) {
        alert('操作失败: ' + error.message)
      } finally {
        this.isLoading = false
      }
    }
  },
  created() {
    // 检查localStorage中的登录状态和token
    if (localStorage.getItem('isLoggedIn') === 'true') {
      this.isLoggedIn = true
      this.token = localStorage.getItem('token') || ''
    }
  }
}
</script>

<style scoped>
.app {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
  font-family: Arial, sans-serif;
}

h1 {
  text-align: center;
  color: #333;
}

.tabs {
  display: flex;
  margin-bottom: 20px;
  border-bottom: 1px solid #ddd;
}

.tabs button {
  padding: 10px 20px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  color: #666;
  border-bottom: 3px solid transparent;
  transition: all 0.3s;
}

.tabs button:hover {
  color: #333;
  background-color: #f5f5f5;
}

.tabs button.active {
  border-bottom-color: #4CAF50;
  color: #4CAF50;
  font-weight: bold;
}

/* 登录相关样式 */
.login-modal {
  z-index: 2000;
}

.login-modal .modal-content {
  max-width: 400px;
}

.error-message {
  color: #f44336;
  margin-top: 10px;
  font-size: 14px;
  text-align: center;
}

.logout-btn {
  margin-left: auto;
  background: linear-gradient(135deg, #ff416c 0%, #ff4b2b 100%);
  color: white;
  padding: 12px 24px;
  border: none;
  border-radius: 25px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 600;
  box-shadow: 0 4px 15px rgba(255, 75, 43, 0.4);
  transition: all 0.3s ease;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.logout-btn:hover {
  background: linear-gradient(135deg, #ff4b2b 0%, #ff416c 100%);
  box-shadow: 0 6px 20px rgba(255, 75, 43, 0.6);
  transform: translateY(-2px);
}

.logout-btn:active {
  transform: translateY(0);
  box-shadow: 0 2px 10px rgba(255, 75, 43, 0.4);
}

.tab-pane {
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.input-with-button {
  display: flex;
  gap: 10px;
}

.input-with-button input {
  flex: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-group input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.btn {
  background-color: #4CAF50;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.3s;
}

.btn:hover {
  background-color: #45a049;
}

.btn:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.delete-btn {
  background-color: #f44336;
}

.delete-btn:hover {
  background-color: #da190b;
}

.select-btn {
  padding: 8px 12px;
  font-size: 14px;
  white-space: nowrap;
}

.add-schedule-btn {
  margin-bottom: 20px;
}

.result {
  margin-top: 20px;
  padding: 15px;
  background-color: #f9f9f9;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.filter-controls {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f5f5f5;
  border-radius: 4px;
}

.filter-controls .form-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-controls select {
  padding: 5px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: center;
}

.page-info {
  margin: 0 10px;
}

.page-size {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-left: 20px;
}

.page-size select {
  padding: 3px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

/* OSS配置卡片样式 */
.oss-config-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.oss-config-card {
  background-color: #f9f9f9;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.3s ease;
}

.oss-config-card:hover {
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.oss-config-info h3 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #333;
  font-size: 18px;
}

.config-detail p {
  margin: 8px 0;
  font-size: 14px;
  color: #555;
}

.oss-config-actions {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}

.oss-config-actions .btn {
  flex: 1;
  padding: 8px 12px;
  font-size: 14px;
}

.result h3 {
  margin-top: 0;
  color: #333;
}

.schedule-list {
  margin-top: 20px;
}

.schedule-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-bottom: 10px;
  background-color: #f9f9f9;
}

.schedule-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.btn.trigger-btn {
  background-color: #4CAF50;
  color: white;
}

.btn.edit-btn {
  background-color: #2196F3;
  color: white;
}

.btn.delete-btn {
  background-color: #f44336;
  color: white;
}

.schedule-info {
  flex: 1;
}

.backup-history-list {
  margin-top: 20px;
}

.backup-record-item {
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-bottom: 10px;
  background-color: #f9f9f9;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.backup-record-info {
  flex: 1;
}

.backup-record-actions {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.btn.restore-btn {
  background-color: #4CAF50;
  color: white;
}

.btn.download-btn {
  background-color: #2196F3;
  color: white;
}

.btn.delete-btn {
  background-color: #f44336;
  color: white;
}

.schedule-info p {
  margin: 5px 0;
}

.no-data {
  text-align: center;
  padding: 40px;
  color: #999;
  font-style: italic;
}

/* 弹窗样式 */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: white;
  padding: 20px;
  border-radius: 4px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.modal-content h3 {
  margin-top: 0;
  color: #333;
  margin-bottom: 15px;
}

.dialog-body {
  margin-bottom: 20px;
}

.dialog-body p {
  margin-bottom: 10px;
  color: #666;
}

.dialog-body input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.modal-footer .btn {
  padding: 8px 16px;
  font-size: 14px;
}

/* 目录选择器样式 */
.directory-modal {
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.directory-selector {
  flex: 1;
  overflow-y: auto;
  max-height: 500px;
}

.directory-tree {
  padding: 10px;
}

.directory-item {
  margin-bottom: 2px;
}

.directory-name {
  padding: 5px 10px;
  cursor: pointer;
  border-radius: 3px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.directory-name:hover {
  background-color: #f0f0f0;
}

.directory-name.selected {
  background-color: #e3f2fd;
  border-left: 3px solid #2196f3;
}

.dir-icon {
  font-size: 12px;
  color: #666;
  width: 12px;
  text-align: center;
}

.file-icon {
  font-size: 14px;
}

.subdirectories {
  margin-left: 20px;
  border-left: 1px solid #ddd;
  padding-left: 10px;
}
</style>
