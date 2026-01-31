---
active: true
iteration: 1
max_iterations: 40
completion_promise: "PRD130_COMPLETE"
started_at: "2026-01-31T01:26:06Z"
---

                                                                                             
  ## PRD-130: Migrate toolindex to tooldiscovery/index                                                      
                                                                                                            
  ### Context                                                                                               
  - Source: /Users/jraymond/Documents/Projects/ApertureStack/toolindex/                                     
  - Target: /Users/jraymond/Documents/Projects/ApertureStack/tooldiscovery/index/                           
  - Working directory: /Users/jraymond/Documents/Projects/ApertureStack/tooldiscovery                       
                                                                                                            
  ### Source Files to Migrate                                                                               
  - index.go (~31KB)                                                                                        
  - index_test.go (~51KB)                                                                                   
  - pagination.go (~1.5KB)                                                                                  
  - contract_test.go (~2KB)                                                                                 
                                                                                                            
  ### Requirements                                                                                          
  1. Copy all .go files from toolindex/ to tooldiscovery/index/                                             
  2. Rename package from 'toolindex' to 'index'                                                             
  3. Update import paths:                                                                                   
     - github.com/jonwraymond/toolindex → github.com/jonwraymond/tooldiscovery/index                        
     - github.com/jonwraymond/toolmodel → github.com/jonwraymond/toolfoundation/model                       
  4. Update go.mod to add toolfoundation dependency                                                         
  5. Create doc.go with package documentation                                                               
  6. All tests must pass with GOWORK=off                                                                    
                                                                                                            
  ### Critical Rules                                                                                        
  - Use GOWORK=off for all go commands                                                                      
  - Package name must be 'index' (not 'toolindex')                                                          
  - Preserve all existing functionality                                                                     
  - Update internal references (e.g., toolindex.X → index.X in test files)                                  
                                                                                                            
  ### Self-Correction                                                                                       
  If tests fail:                                                                                            
  1. Read the error message carefully                                                                       
  2. Check for missed import path updates                                                                   
  3. Check for missed package name references                                                               
  4. Fix and re-run tests                                                                                   
                                                                                                            
  If build fails:                                                                                           
  1. Check go.mod has correct dependencies                                                                  
  2. Run 'GOWORK=off go mod tidy'                                                                           
  3. Verify toolfoundation/model exports required types                                                     
                                                                                                            
  ### Verification Commands                                                                                 
  ```bash                                                                                                
  cd /Users/jraymond/Documents/Projects/ApertureStack/tooldiscovery                                         
  GOWORK=off go build ./index/...                                                                           
  GOWORK=off go test ./index/... -v                                                                         
  GOWORK=off go test ./index/... -cover                                                                     
  ```                                                                                                    
                                                                                                            
  ### Completion Criteria                                                                                   
  ALL of these must be true:                                                                                
  - [ ] index/ directory exists with migrated files                                                         
  - [ ] Package declaration is 'package index'                                                              
  - [ ] All imports use jonwraymond/tooldiscovery/index                                                     
  - [ ] All imports use jonwraymond/toolfoundation/model                                                    
  - [ ] GOWORK=off go build ./index/... succeeds                                                            
  - [ ] GOWORK=off go test ./index/... passes                                                               
  - [ ] Coverage >70%                                                                                       
  - [ ] doc.go exists with package documentation                                                            
  - [ ] Committed with message: feat(index): migrate toolindex package                                      
  - [ ] Pushed to origin/main                                                                               
                                                                                                            
  ### Commit Message Template                                                                               
  feat(index): migrate toolindex package                                                                    
                                                                                                            
  Migrate the tool registry from standalone toolindex repository.                                           
                                                                                                            
  Package contents:                                                                                         
  - Index interface with CRUD operations                                                                    
  - InMemoryIndex for ephemeral storage                                                                     
  - FileIndex for persistent storage                                                                        
  - Searcher interface for pluggable search                                                                 
  - Progressive disclosure support                                                                          
  - Pagination support                                                                                      
                                                                                                            
  Dependencies:                                                                                             
  - github.com/jonwraymond/toolfoundation/model                                                             
                                                                                                            
  Migration: github.com/jonwraymond/toolindex → tooldiscovery/index                                         
                                                                                                            
  Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>                                                   
                                                                                                            
  ### Completion Signal                                                                                     
  When ALL criteria verified, output:                                                                       
  <promise>PRD130_COMPLETE</promise>                                                                        
  
