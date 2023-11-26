# internal

## Changelog

### Iteration 10

It this iteration I should add 500 error return if database is not connected.

`gin.recovery()` usually used for catch panic and return 500 instead. But I can't use it, because initialize is not handler of gin. So, I add custom check. I think it's not best pracite case, but I don't know any other solution.

### Iteration 7

In this iteration I changed main logic of routes.

Before I use 3 gin groups:
1. `/app` - Check live and show all collected values
2. `/update` - Update metrics in local database
3. `/value` - Get one value from database

So, this method have one disadvatage, `/app` and `/value` return data, but use one database together. And all endpoints have one structure for metrics. In previos iterations it not was a problem, but now we have JSON structure for our metrics and copy this structure in 3+ folders can cause troubles.

Now, I use only one endpoint (`/app`) and I shouldn't copy code in other folders. This solution reduces the number of folders in the project and, most importantly, the amount of code.