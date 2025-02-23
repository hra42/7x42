# **Phase 2: Image Generation (Week 2)**

### Day 1-2: Image Generation Infrastructure
1. **Model Extensions**
   ```go
   // models/image.go
   type Image struct {
       gorm.Model
       Prompt        string
       NegativePrompt string `gorm:"default:''"`
       Status        string    // pending/processing/completed/failed
       URL           string    // stored image URL
       ChatID        uint      // associated chat
       MessageID     uint      // associated message
       Width         int       `gorm:"default:512"`
       Height        int       `gorm:"default:512"`
       Model         string    // which model to use
       Seed          int64     // for reproducibility
   }
   ```

2. **Storage Setup**
- Either MiniIO for local storage
- Or external S3 Storage

### Day 3-4: Replikate Image API Integration
1. **Image Generation Service**
    - Queue management for image requests
    - Progress tracking
    - Error handling and retries
    - Result storage in MinIO

2. **WebSocket Extensions**
   ```go
   type WSImageMessage struct {
       Type          string `json:"type"` // image_request/image_progress/image_complete
       Prompt        string `json:"prompt,omitempty"`
       Progress      int    `json:"progress,omitempty"`
       ImageURL      string `json:"image_url,omitempty"`
       Error         string `json:"error,omitempty"`
   }
   ```

### Day 5-6: UI Implementation
1. **Image Generation UI**
   ```html
   <!-- templates/image-generator.html -->
   <div x-data="imageGenerator" class="dark:bg-gray-800 p-4 rounded-lg">
     <!-- Image Generation Form -->
     <form @submit.prevent="generateImage" class="space-y-4">
       <div>
         <label class="block text-sm font-medium dark:text-gray-200">Prompt</label>
         <textarea 
           x-model="prompt" 
           class="w-full dark:bg-gray-700 rounded"
           rows="3"></textarea>
       </div>
       
       <!-- Advanced Options (collapsible) -->
       <div x-show="showAdvanced">
         <div class="grid grid-cols-2 gap-4">
           <div>
             <label>Width</label>
             <input type="number" x-model="width" step="64" min="384" max="1024">
           </div>
           <div>
             <label>Height</label>
             <input type="number" x-model="height" step="64" min="384" max="1024">
           </div>
         </div>
       </div>

       <!-- Progress Bar -->
       <div x-show="isGenerating" class="relative pt-1">
         <div class="overflow-hidden h-2 text-xs flex rounded bg-gray-700">
           <div 
             :style="`width: ${progress}%`"
             class="animate-pulse shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-blue-500">
           </div>
         </div>
       </div>

       <button 
         type="submit" 
         :disabled="isGenerating"
         class="w-full btn-primary">
         Generate Image
       </button>
     </form>

     <!-- Results Gallery -->
     <div class="grid grid-cols-2 gap-4 mt-6">
       <template x-for="image in generatedImages">
         <div class="relative group">
           <img :src="image.url" class="rounded-lg">
           <div class="absolute bottom-0 p-2 bg-black/50 w-full">
             <p class="text-xs text-white" x-text="image.prompt"></p>
           </div>
         </div>
       </template>
     </div>
   </div>
   ```

2. **Alpine.js Logic**
   ```javascript
   document.addEventListener('alpine:init', () => {
     Alpine.data('imageGenerator', () => ({
       prompt: '',
       width: 512,
       height: 512,
       showAdvanced: false,
       isGenerating: false,
       progress: 0,
       generatedImages: [],
       
       async generateImage() {
         this.isGenerating = true;
         this.progress = 0;
         
         // Send via WebSocket
         window.ws.send(JSON.stringify({
           type: 'image_request',
           prompt: this.prompt,
           width: this.width,
           height: this.height
         }));
       },

       handleWSMessage(msg) {
         if (msg.type === 'image_progress') {
           this.progress = msg.progress;
         } else if (msg.type === 'image_complete') {
           this.isGenerating = false;
           this.generatedImages.unshift({
             url: msg.image_url,
             prompt: this.prompt
           });
         }
       }
     }))
   })
   ```

### Day 7: Integration & Testing
1. **Integration Features**
    - Image generation from chat context
    - Image variation generation
    - Prompt history/favorites

2. **Testing & Optimization**
    - Image generation queue performance
    - Storage optimization
    - Error recovery scenarios

### Key Deliverables for Phase 2:
1. **Image Generation System:**
    - Prompt → Image pipeline
    - Progress tracking
    - Image storage & retrieval
    - Queue management

2. **Enhanced UI:**
    - Image generation form
    - Real-time progress updates
    - Image gallery view
    - Advanced options panel

3. **Integration with Chat:**
    - Seamless chat-to-image generation
    - Image results in chat history
    - Shared prompt context