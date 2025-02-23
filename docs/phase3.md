# **Phase 3: Security & User Management (Week 3)**

### Day 1-2: User Authentication System
1. **User Model & Database**
   ```go
   // models/user.go
   type User struct {
       gorm.Model
       Email        string `gorm:"uniqueIndex"`
       Password     string `gorm:"-"` // not stored, only for validation
       PasswordHash string
       FirstName    string
       LastName     string
       Role         string `gorm:"default:'user'"` // user/admin
       ApiKey       string `gorm:"uniqueIndex"`
       Credits      int    `gorm:"default:100"`
       LastLogin    time.Time
       Active       bool `gorm:"default:true"`
       
       // Relations
       Chats        []Chat
       Images       []Image
   }

   // Add foreign keys to existing models
   type Chat struct {
       gorm.Model
       UserID    uint
       User      User
       // ... existing fields
   }

   type Image struct {
       gorm.Model
       UserID    uint
       User      User
       // ... existing fields
   }
   ```

2. **JWT Authentication Middleware**
   ```go
   // middleware/auth.go
   func JWTMiddleware() fiber.Handler {
       return jwtware.New(jwtware.Config{
           SigningKey:    []byte(os.Getenv("JWT_SECRET")),
           ErrorHandler: func(c *fiber.Ctx, err error) error {
               return c.Redirect("/login")
           },
       })
   }

   // Custom WebSocket auth
   func WSAuthMiddleware() func(*websocket.Conn) bool {
       return func(conn *websocket.Conn) bool {
           token := conn.Query("token")
           // Validate JWT token
           return validateToken(token)
       }
   }
   ```

### Day 3-4: User Management UI & API
1. **Authentication Routes**
   ```go
   // routes/auth.go
   app.Post("/api/register", handlers.Register)
   app.Post("/api/login", handlers.Login)
   app.Post("/api/logout", handlers.Logout)
   app.Get("/api/me", middleware.Protected(), handlers.GetProfile)
   app.Put("/api/me", middleware.Protected(), handlers.UpdateProfile)
   ```

2. **User Management UI**
   ```html
   <!-- templates/profile.html -->
   <div x-data="userProfile" class="container mx-auto p-4">
     <!-- Profile Section -->
     <div class="dark:bg-gray-800 rounded-lg p-6 mb-6">
       <h2 class="text-2xl mb-4">Profile Settings</h2>
       
       <form @submit.prevent="updateProfile" class="space-y-4">
         <div class="grid grid-cols-2 gap-4">
           <div>
             <label>First Name</label>
             <input type="text" x-model="user.firstName">
           </div>
           <div>
             <label>Last Name</label>
             <input type="text" x-model="user.lastName">
           </div>
         </div>

         <div>
           <label>Email</label>
           <input type="email" x-model="user.email" disabled>
         </div>

         <div>
           <label>API Key</label>
           <div class="flex">
             <input type="text" x-model="user.apiKey" readonly>
             <button @click="regenerateApiKey" class="ml-2">
               Regenerate
             </button>
           </div>
         </div>

         <button type="submit" class="btn-primary">
           Save Changes
         </button>
       </form>
     </div>

     <!-- Usage Stats -->
     <div class="dark:bg-gray-800 rounded-lg p-6">
       <h3 class="text-xl mb-4">Usage</h3>
       <div class="grid grid-cols-3 gap-4">
         <div>
           <span class="block text-sm">Credits</span>
           <span class="text-2xl" x-text="user.credits"></span>
         </div>
         <div>
           <span class="block text-sm">Chats</span>
           <span class="text-2xl" x-text="stats.totalChats"></span>
         </div>
         <div>
           <span class="block text-sm">Images</span>
           <span class="text-2xl" x-text="stats.totalImages"></span>
         </div>
       </div>
     </div>
   </div>
   ```

### Day 5: Rate Limiting & Usage Tracking
1. **Rate Limiting Middleware**
   ```go
   // middleware/ratelimit.go
   func RateLimit() fiber.Handler {
       return limiter.New(limiter.Config{
           Max:        60,
           Expiration: 1 * time.Minute,
           KeyGenerator: func(c *fiber.Ctx) string {
               // Use user ID if authenticated, IP otherwise
               user := c.Locals("user")
               if user != nil {
                   return fmt.Sprintf("user:%d", user.(*models.User).ID)
               }
               return c.IP()
           },
           LimitReached: func(c *fiber.Ctx) error {
               return c.Status(429).JSON(fiber.Map{
                   "error": "Too many requests",
               })
           },
       })
   }
   ```

2. **Credit System**
   ```go
   // services/credits.go
   type CreditManager struct {
       db *gorm.DB
   }

   func (cm *CreditManager) DeductCredits(userID uint, amount int) error {
       return cm.db.Transaction(func(tx *gorm.DB) error {
           var user models.User
           if err := tx.Lock().First(&user, userID).Error; err != nil {
               return err
           }
           
           if user.Credits < amount {
               return errors.New("insufficient credits")
           }
           
           user.Credits -= amount
           return tx.Save(&user).Error
       })
   }
   ```

### Day 6: Security Hardening
1. **Security Headers Middleware**
   ```go
   app.Use(helmet.New())
   app.Use(csrf.New())
   ```

2. **Input Validation**
   ```go
   // validators/user.go
   type RegisterInput struct {
       Email     string `validate:"required,email"`
       Password  string `validate:"required,min=8"`
       FirstName string `validate:"required"`
       LastName  string `validate:"required"`
   }

   func ValidateRegisterInput(input RegisterInput) error {
       validate := validator.New()
       return validate.Struct(input)
   }
   ```

### Day 7: Testing & Monitoring
1. **Security Tests**
   ```go
   // tests/security_test.go
   func TestJWTProtection(t *testing.T) {
       // Test protected routes
       // Test token expiration
       // Test invalid tokens
   }
   ```

2. **Monitoring Setup**
   ```go
   // monitoring/metrics.go
   var (
       activeUsers = prometheus.NewGauge(prometheus.GaugeOpts{
           Name: "active_users",
           Help: "Number of currently active users",
       })
       
       creditUsage = prometheus.NewCounterVec(
           prometheus.CounterOpts{
               Name: "credit_usage_total",
               Help: "Total credits used by service",
           },
           []string{"service_type"},
       )
   )
   ```

### Key Deliverables for Phase 3:
1. **Authentication System:**
    - User registration/login
    - JWT-based session management
    - Password reset flow
    - API key management

2. **Security Features:**
    - Rate limiting
    - Input validation
    - CSRF protection
    - Security headers

3. **User Management:**
    - Profile management
    - Usage tracking
    - Credit system
    - Admin interface
