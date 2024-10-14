# gomongo
Gomongo is a very thin wrapper to mongo go driver. If increase productivity and code reliablity by using generics to represent a document. Also this wrapper respond with a single structure that holds all the inforamtion you need. 

Each mongo function can be executed as sync or go function. In case you are using go function, this module returns a none blocking channel. You can drain the channel whenever you like to get the result of the operation. This is good for starting several actions and then collect the result down the code.

# Installing
use the following command to add gomongo to your project
``` bash

go get github.com/sagiforbes/gomongo


```


# Usage examples

You start by creating a client struct to be used by all gomongo functions. In this example we run mongo as a single node on localhost. To get a client run

``` go



```




