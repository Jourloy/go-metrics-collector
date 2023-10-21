# internal

## Changelog

### Iteration 7

In this iteration I changed main logic of routes.

Before I use 3 gin groups:
1. `/app` - Check live and show all collected values
2. `/update` - Update metrics in local database
3. `/value` - Get one value from database

So, this method have one disadvatage, `/app` and `/value` return data, but use one database together. And all endpoints have one structure for metrics. In previos iterations it not was a problem, but now we have JSON structure for our metrics and copy this structure in 3+ folders can cause troubles.

Now, I use only one endpoint (`/app`) and I shouldn't copy code in other folders. This solution reduces the number of folders in the project and, most importantly, the amount of code.