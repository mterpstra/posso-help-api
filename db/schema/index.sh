# This creates an index that forces tag and account to be unique.
# In other words, duplicate tags per account are not allowed. 
# EXCEPT: Except when tag is zero, then we allow it.
db.births.createIndex(                                                                                                            
  { account: 1, tag: 1 },                                                                                                         
  {                                                                                                                               
    unique: true,                                                                                                                 
    partialFilterExpression: { tag: { $gt: 0 } }                                                                                  
  }                                                                                                                               
)  
