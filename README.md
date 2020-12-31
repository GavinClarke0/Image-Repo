Image Repo is basic token authicated image non distributed image repo written in pure golang including dependancies.

**Features**

* Single and bulk zip file upload 
* bulk and single image retrieval  
* Multitentant token based authentication and private files 
* Retrieved imaged server side transformations 
* Super Fast upload/retreival thanks to PebbleDB for metadata storage and file data stored on local file system 

** Examples **

1. To use image-repo autheticate your client by retrieving a token via "/login" 

2. Upload images via "/image/putimage" with the body containg the image and the content0-type equal to the image type or bulk upload via "/image/putimages"

2.1 mark images private with "?visibility=private" to ensure only the owner client can retrieve the image 

3. Revtrieve images via "/image/getimage/{imageId}" with the ability to abritary transform the image server side query params such as "width=900" 

4. New tentants can be created "/createuser"

**Design Considerations**

Application utalizes a decoupled storage engine (currently local) top allow easy swapping for cloud storage file storage such as AWS s3 

**TODO** 

* Docker Image 
* Visual Demo  
