---
active: true
iteration: 1
max_iterations: 30
completion_promise: "PRD121_COMPLETE"
started_at: "2026-01-31T01:06:32Z"
---

                                                                                             
  ## Task: Execute PRD-121 - Migrate tooladapter to toolfoundation/adapter                                  
                                                                                                            
  ### Description                                                                                          
  Migrate the tooladapter package from standalone repo (jonwraymond/tooladapter) to consolidated repo       
  (jonwraymond/toolfoundation/adapter).         )                                                           
                                                                                                            
  ### Source & Target                                                                                       
  - Source: /Users/jraymond/Documents/Projects/ApertureStack/tooladapter/                                   
  - Target: /Users/jraymond/Documents/Projects/ApertureStack/toolfoundation/adapter/                        
                                                                                                            
  ### Requirements                                                                                          
  - [ ] Copy all .go files from tooladapter to toolfoundation/adapter/                                      
  - [ ] Update package declaration from 'package tooladapter' to 'package adapter'                          
  - [ ] Update go.mod to include any new dependencies from tooladapter                                      
  - [ ] Update internal import paths to github.com/jonwraymond/toolfoundation/adapter                       
  - [ ] If adapter imports toolmodel, update to github.com/jonwraymond/toolfoundation/model                 
  - [ ] Ensure all tests pass with 'go test ./adapter/...'                                                  
  - [ ] Ensure 'go build ./...' succeeds                                                                    
  - [ ] Commit changes with proper message                                                                  
                                                                                                            
  ### Execution Steps                                                                                       
  1. Create adapter/ directory in toolfoundation                                                            
  2. Copy all *.go files from tooladapter to toolfoundation/adapter/                                        
  3. Update package declarations (tooladapter -> adapter)                                                   
  4. Update any internal imports (tooladapter -> adapter, toolmodel -> model)                               
  5. Merge any new dependencies into toolfoundation/go.mod                                                  
  6. Run go mod tidy                                                                                        
  7. Run go build ./...                                                                                     
  8. Run go test ./adapter/...                                                                              
  9. Commit and push                                                                                        
                                                                                                            
  ### Self-Correction                                                                                       
  If go build fails:                                                                                        
  1. Read the error message                                                                                 
  2. Check for missing imports or wrong package names                                                       
  3. If toolmodel import fails, ensure it points to toolfoundation/model                                    
  4. Fix and retry                                                                                          
                                                                                                            
  If go test fails:                                                                                         
  1. Read test output carefully                                                                             
  2. Check if test imports need updating                                                                    
  3. Fix failing tests                                                                                      
                                                                                                            
  If import cycle detected:                                                                                 
  1. Check for circular dependencies between model and adapter                                              
  2. May need to reorganize code                                                                            
                                                                                                            
  If stuck for 5+ iterations on same issue:                                                                 
  1. Document the blocker                                                                                   
  2. List what was attempted                                                                                
  3. Output: <promise>NEEDS_HELP</promise>                                                                  
                                                                                                            
  ### Verification Commands                                                                                 
  ```bash                                                                                                
  # Verify package exists                                                                                   
  ls toolfoundation/adapter/*.go                                                                            
                                                                                                            
  # Verify build                                                                                            
  cd toolfoundation && GOWORK=off go build ./...                                                            
                                                                                                            
  # Verify tests                                                                                            
  cd toolfoundation && GOWORK=off go test ./adapter/... -v                                                  
                                                                                                            
  # Verify package name                                                                                     
  grep '^package adapter' toolfoundation/adapter/*.go | head -3                                             
                                                                                                            
  # Verify no old imports                                                                                   
  grep -r 'tooladapter' toolfoundation/adapter/ || echo 'No old imports found'                              
  grep -r 'toolmodel' toolfoundation/adapter/ || echo 'No toolmodel imports found'                          
  ```                                                                                                    
                                                                                                            
  ### Completion Criteria                                                                                   
  ALL of the following must be true:                                                                        
  1. toolfoundation/adapter/ directory exists with .go files                                                
  2. All files have 'package adapter' declaration                                                           
  3. No references to old import paths (tooladapter, toolmodel)                                             
  4. `go build ./...` succeeds in toolfoundation                                                          
  5. `go test ./adapter/...` passes                                                                       
  6. Changes committed and pushed                                                                           
                                                                                                            
  When ALL criteria verified:                                                                               
  Output: <promise>PRD121_COMPLETE</promise>                                                                
  
