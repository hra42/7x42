# **Phase 4: Agent & Workflow System (Week 4)**

### Day 1-2: Workflow Engine Core
1. **Workflow Models**
   ```go
   // models/workflow.go
   type Workflow struct {
       gorm.Model
       Name        string
       UserID      uint
       User        User
       IsActive    bool `gorm:"default:true"`
       Trigger     string    // message/schedule/api
       Steps       []WorkflowStep
       Variables   datatypes.JSON
   }

   type WorkflowStep struct {
       gorm.Model
       WorkflowID  uint
       Order       int
       Type        string    // chat/image/condition/api
       Config      datatypes.JSON
       RetryPolicy *RetryConfig
   }

   type WorkflowExecution struct {
       gorm.Model
       WorkflowID  uint
       Status      string    // running/completed/failed
       StartedAt   time.Time
       EndedAt     *time.Time
       Results     []WorkflowStepResult
       Variables   datatypes.JSON
   }
   ```

2. **Workflow Engine**
   ```go
   // services/workflow/engine.go
   type Engine struct {
       db          *gorm.DB
       queue       *asynq.Client
       openrouter  *openrouter.Client
   }

   func (e *Engine) ExecuteWorkflow(ctx context.Context, workflowID uint, input map[string]interface{}) error {
       execution := &models.WorkflowExecution{
           WorkflowID: workflowID,
           Status:    "running",
           StartedAt: time.Now(),
       }

       // Start transaction
       return e.db.Transaction(func(tx *gorm.DB) error {
           if err := tx.Create(execution).Error; err != nil {
               return err
           }

           // Queue first step
           task := asynq.NewTask(
               "workflow:step",
               map[string]interface{}{
                   "execution_id": execution.ID,
                   "step_order":  1,
                   "input":       input,
               },
           )
           
           _, err := e.queue.EnqueueContext(ctx, task)
           return err
       })
   }
   ```

### Day 3-4: Agent System
1. **Agent Definition**
   ```go
   // models/agent.go
   type Agent struct {
       gorm.Model
       Name        string
       UserID      uint
       User        User
       SystemPrompt string    // base personality/instructions
       Memory      []AgentMemory
       Functions   []AgentFunction
       Model       string    // which LLM to use
   }

   type AgentMemory struct {
       gorm.Model
       AgentID     uint
       Type        string    // conversation/fact/learned
       Content     string
       Embedding   vector.Vector `gorm:"type:vector(1536)"`
       LastAccessed time.Time
   }

   type AgentFunction struct {
       gorm.Model
       AgentID     uint
       Name        string
       Description string
       Parameters  datatypes.JSON
       WorkflowID  *uint     // optional linked workflow
   }
   ```

2. **Agent Service**
   ```go
   // services/agent/service.go
   type AgentService struct {
       db          *gorm.DB
       openrouter  *openrouter.Client
       vectorStore *pgvector.Store
   }

   func (s *AgentService) ProcessMessage(ctx context.Context, agentID uint, input string) (*ChatResponse, error) {
       agent, err := s.loadAgent(agentID)
       if err != nil {
           return nil, err
       }

       // Retrieve relevant memories
       memories := s.searchMemories(agent, input)
       
       // Build context from memories
       context := buildContext(agent.SystemPrompt, memories)
       
       // Process with OpenRouter
       response, err := s.openrouter.ChatCompletion(ctx, &openrouter.ChatRequest{
           Model: agent.Model,
           Messages: []openrouter.Message{
               {Role: "system", Content: context},
               {Role: "user", Content: input},
           },
           Functions: s.buildFunctionDefinitions(agent),
       })

       // Handle function calls if any
       if response.FunctionCall != nil {
           return s.handleFunctionCall(ctx, agent, response.FunctionCall)
       }

       return response, nil
   }
   ```

### Day 5: Workflow & Agent UI
1. **Workflow Builder UI**
   ```html
   <!-- templates/workflow-builder.html -->
   <div x-data="workflowBuilder" class="h-screen flex">
     <!-- Steps Palette -->
     <div class="w-64 border-r dark:border-gray-700 p-4">
       <h3 class="text-lg mb-4">Steps</h3>
       <div class="space-y-2">
         <template x-for="step in availableSteps">
           <div 
             class="p-2 border dark:border-gray-600 rounded cursor-move"
             draggable="true"
             @dragstart="dragStart($event, step)">
             <span x-text="step.name"></span>
           </div>
         </template>
       </div>
     </div>

     <!-- Workflow Canvas -->
     <div class="flex-1 p-4">
       <div class="flex justify-between mb-4">
         <input 
           type="text" 
           x-model="workflow.name" 
           class="text-xl bg-transparent"
           placeholder="Workflow Name">
         
         <button 
           @click="saveWorkflow"
           class="btn-primary">
           Save Workflow
         </button>
       </div>

       <!-- Steps Container -->
       <div 
         class="space-y-4"
         @dragover.prevent
         @drop="dropStep($event)">
         <template x-for="(step, index) in workflow.steps" :key="step.id">
           <div class="border dark:border-gray-600 p-4 rounded">
             <!-- Step Configuration -->
             <div class="flex justify-between mb-2">
               <span x-text="step.type"></span>
               <button @click="removeStep(index)">&times;</button>
             </div>
             
             <!-- Dynamic Step Config -->
             <div x-show="step.type === 'chat'">
               <textarea 
                 x-model="step.config.prompt"
                 placeholder="Enter prompt..."
                 class="w-full"></textarea>
             </div>
             
             <div x-show="step.type === 'condition'">
               <select x-model="step.config.operator">
                 <option>contains</option>
                 <option>equals</option>
                 <option>greater_than</option>
               </select>
               <input 
                 type="text" 
                 x-model="step.config.value"
                 placeholder="Compare value">
             </div>
           </div>
         </template>
       </div>
     </div>
   </div>
   ```

2. **Agent Management UI**
   ```html
   <!-- templates/agent-manager.html -->
   <div x-data="agentManager" class="container mx-auto p-4">
     <!-- Agent List -->
     <div class="grid grid-cols-3 gap-4 mb-8">
       <template x-for="agent in agents">
         <div class="dark:bg-gray-800 p-4 rounded-lg">
           <h3 x-text="agent.name" class="text-xl mb-2"></h3>
           <p class="text-sm mb-4" x-text="agent.description"></p>
           
           <div class="flex justify-between">
             <button 
               @click="editAgent(agent)"
               class="btn-secondary">
               Edit
             </button>
             <button 
               @click="chatWithAgent(agent)"
               class="btn-primary">
               Chat
             </button>
           </div>
         </div>
       </template>
     </div>

     <!-- Agent Editor Modal -->
     <div x-show="showEditor" class="modal">
       <div class="modal-content">
         <h2 class="text-2xl mb-4">Configure Agent</h2>
         
         <form @submit.prevent="saveAgent">
           <div class="space-y-4">
             <div>
               <label>Name</label>
               <input type="text" x-model="editingAgent.name">
             </div>
             
             <div>
               <label>System Prompt</label>
               <textarea 
                 x-model="editingAgent.systemPrompt"
                 rows="4"></textarea>
             </div>
             
             <div>
               <label>Model</label>
               <select x-model="editingAgent.model">
                 <option value="gpt-4">GPT-4</option>
                 <option value="claude-2">Claude 2</option>
               </select>
             </div>

             <!-- Functions -->
             <div>
               <h4 class="text-lg mb-2">Functions</h4>
               <template x-for="(func, index) in editingAgent.functions">
                 <div class="border dark:border-gray-600 p-2 mb-2">
                   <input 
                     type="text" 
                     x-model="func.name"
                     placeholder="Function name">
                   <select x-model="func.workflowId">
                     <option value="">Select workflow...</option>
                     <template x-for="wf in workflows">
                       <option :value="wf.id" x-text="wf.name"></option>
                     </template>
                   </select>
                 </div>
               </template>
               <button 
                 @click="addFunction"
                 type="button"
                 class="btn-secondary">
                 Add Function
               </button>
             </div>
           </div>

           <div class="flex justify-end mt-4">
             <button type="submit" class="btn-primary">
               Save Agent
             </button>
           </div>
         </form>
       </div>
     </div>
   </div>
   ```

### Day 6-7: Integration & Testing
1. **Integration Tests**
   ```go
   // tests/workflow_test.go
   func TestWorkflowExecution(t *testing.T) {
       // Test basic workflow execution
       // Test condition branching
       // Test error handling and retries
       // Test variable passing between steps
   }

   // tests/agent_test.go
   func TestAgentFunctions(t *testing.T) {
       // Test function calling
       // Test memory retrieval
       // Test workflow integration
   }
   ```

2. **Monitoring & Debugging**
   ```go
   // monitoring/workflow.go
   var (
       workflowExecutions = prometheus.NewCounterVec(
           prometheus.CounterOpts{
               Name: "workflow_executions_total",
               Help: "Number of workflow executions",
           },
           []string{"workflow", "status"},
       )

       stepDuration = prometheus.NewHistogramVec(
           prometheus.HistogramOpts{
               Name: "workflow_step_duration_seconds",
               Help: "Duration of workflow steps",
           },
           []string{"workflow", "step_type"},
       )
   )
   ```

### Key Deliverables for Phase 4:
1. **Workflow System:**
    - Visual workflow builder
    - Step library (chat, image, conditions)
    - Execution engine
    - Variable handling

2. **Agent System:**
    - Agent configuration
    - Memory management
    - Function calling
    - Workflow integration

3. **Integration Features:**
    - Agent-to-workflow bridging
    - Shared variable context
    - Error handling
    - Monitoring
