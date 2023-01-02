"use strict";(self.webpackChunkcompass=self.webpackChunkcompass||[]).push([[115],{3905:function(e,t,n){n.d(t,{Zo:function(){return u},kt:function(){return d}});var a=n(7294);function s(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function r(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,a)}return n}function o(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?r(Object(n),!0).forEach((function(t){s(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):r(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function i(e,t){if(null==e)return{};var n,a,s=function(e,t){if(null==e)return{};var n,a,s={},r=Object.keys(e);for(a=0;a<r.length;a++)n=r[a],t.indexOf(n)>=0||(s[n]=e[n]);return s}(e,t);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(a=0;a<r.length;a++)n=r[a],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(s[n]=e[n])}return s}var p=a.createContext({}),l=function(e){var t=a.useContext(p),n=t;return e&&(n="function"==typeof e?e(t):o(o({},t),e)),n},u=function(e){var t=l(e.components);return a.createElement(p.Provider,{value:t},e.children)},m={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},c=a.forwardRef((function(e,t){var n=e.components,s=e.mdxType,r=e.originalType,p=e.parentName,u=i(e,["components","mdxType","originalType","parentName"]),c=l(n),d=s,h=c["".concat(p,".").concat(d)]||c[d]||m[d]||r;return n?a.createElement(h,o(o({ref:t},u),{},{components:n})):a.createElement(h,o({ref:t},u))}));function d(e,t){var n=arguments,s=t&&t.mdxType;if("string"==typeof e||s){var r=n.length,o=new Array(r);o[0]=c;var i={};for(var p in t)hasOwnProperty.call(t,p)&&(i[p]=t[p]);i.originalType=e,i.mdxType="string"==typeof e?e:s,o[1]=i;for(var l=2;l<r;l++)o[l]=n[l];return a.createElement.apply(null,o)}return a.createElement.apply(null,n)}c.displayName="MDXCreateElement"},855:function(e,t,n){n.r(t),n.d(t,{assets:function(){return u},contentTitle:function(){return p},default:function(){return d},frontMatter:function(){return i},metadata:function(){return l},toc:function(){return m}});var a=n(7462),s=n(3366),r=(n(7294),n(3905)),o=["components"],i={},p="1. My First Asset",l={unversionedId:"tour/my-first-asset",id:"tour/my-first-asset",title:"1. My First Asset",description:"Before starting the tour, make sure you have a running Compass instance. You can refer this installation guide.",source:"@site/docs/tour/1-my-first-asset.md",sourceDirName:"tour",slug:"/tour/my-first-asset",permalink:"/compass/tour/my-first-asset",draft:!1,editUrl:"https://github.com/odpf/compass/edit/master/docs/docs/tour/1-my-first-asset.md",tags:[],version:"current",sidebarPosition:1,frontMatter:{},sidebar:"docsSidebar",previous:{title:"Installation",permalink:"/compass/installation"},next:{title:"2. Querying your Assets",permalink:"/compass/tour/querying-assets"}},u={},m=[{value:"1.1 Introduction",id:"11-introduction",level:2},{value:"1.2 Hello, <del>World</del> Asset!",id:"12-hello-world-asset",level:2},{value:"1.3 Sending your first asset to Compass",id:"13-sending-your-first-asset-to-compass",level:2},{value:"Conclusion",id:"conclusion",level:2}],c={toc:m};function d(e){var t=e.components,n=(0,s.Z)(e,o);return(0,r.kt)("wrapper",(0,a.Z)({},c,n,{components:t,mdxType:"MDXLayout"}),(0,r.kt)("h1",{id:"1-my-first-asset"},"1. My First Asset"),(0,r.kt)("p",null,"Before starting the tour, make sure you have a running Compass instance. You can refer this ",(0,r.kt)("a",{parentName:"p",href:"../installation"},"installation guide"),"."),(0,r.kt)("h2",{id:"11-introduction"},"1.1 Introduction"),(0,r.kt)("p",null,"In Compass, we call every metadata that you input as an ",(0,r.kt)("a",{parentName:"p",href:"../concepts/asset"},"Asset"),". All your tables, dashboards, topics, jobs are an example of assets."),(0,r.kt)("p",null,"In this section, we will help you to build your first Asset and hopefully it will give your clear idea about what an Asset is in Compass."),(0,r.kt)("h2",{id:"12-hello-world-asset"},"1.2 Hello, ",(0,r.kt)("del",{parentName:"h2"},"World")," Asset!"),(0,r.kt)("p",null,"Let's imagine we have a ",(0,r.kt)("inlineCode",{parentName:"p"},"postgres")," instance that we keep referring to as our ",(0,r.kt)("inlineCode",{parentName:"p"},"main-postgres"),". Inside it there is a database called ",(0,r.kt)("inlineCode",{parentName:"p"},"my-database")," that has plenty of tables. One of the tables is named ",(0,r.kt)("inlineCode",{parentName:"p"},"orders"),", and below is how you represent that ",(0,r.kt)("inlineCode",{parentName:"p"},"table")," as an Compass' Asset."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-json"},'{\n  "urn": "main-postgres:my-database.orders",\n  "type": "table",\n  "service": "postgres",\n  "name": "orders",\n  "data": {\n    "database": "my-database",\n    "namespace": "main-postgres"\n  }\n}\n')),(0,r.kt)("ul",null,(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("p",{parentName:"li"},(0,r.kt)("strong",{parentName:"p"},"urn")," is a unique name you assign to an asset. You need to make sure you don't have a duplicate urns across all of your assets because Compass treats ",(0,r.kt)("inlineCode",{parentName:"p"},"urn")," as an identifier of your asset. For this example, we use the following format to make sure our urn is unique, ",(0,r.kt)("inlineCode",{parentName:"p"},"{NAMESPACE}:{DB_NAME}.{TABLE_NAME}"),".")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("p",{parentName:"li"},(0,r.kt)("strong",{parentName:"p"},"type")," is your Asset's type. The value for type has to be recognizable by Compass. More info about Asset's Type can be found ",(0,r.kt)("a",{parentName:"p",href:"../concepts/type"},"here"),".")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("p",{parentName:"li"},(0,r.kt)("strong",{parentName:"p"},"service")," can be seen as the source of your asset. ",(0,r.kt)("inlineCode",{parentName:"p"},"service")," can be anything, in this case since our ",(0,r.kt)("inlineCode",{parentName:"p"},"orders")," table resides in ",(0,r.kt)("inlineCode",{parentName:"p"},"postgres"),", we can just put ",(0,r.kt)("inlineCode",{parentName:"p"},"postgres")," as the service.")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("p",{parentName:"li"},(0,r.kt)("strong",{parentName:"p"},"name")," is the name of your asset, it does not have to be unique. We don't need to worry to get mixed up if there are other tables with the same name, ",(0,r.kt)("inlineCode",{parentName:"p"},"urn")," will be the main identifier for your asset, that is why we need to make it unique across all of your assets.")),(0,r.kt)("li",{parentName:"ul"},(0,r.kt)("p",{parentName:"li"},(0,r.kt)("strong",{parentName:"p"},"data")," can hold your asset's extra details if there is any. In the example, we use it to store information of the ",(0,r.kt)("strong",{parentName:"p"},"database name")," and the ",(0,r.kt)("strong",{parentName:"p"},"alias/namespace")," that we use when referring the postgres instance."))),(0,r.kt)("h2",{id:"13-sending-your-first-asset-to-compass"},"1.3 Sending your first asset to Compass"),(0,r.kt)("p",null,"Here is the asset that we built on previous section."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-json"},'{\n  "urn": "main-postgres:my-database.orders",\n  "type": "table",\n  "service": "postgres",\n  "name": "orders",\n  "data": {\n    "database": "my-database",\n    "namespace": "main-postgres"\n  }\n}\n')),(0,r.kt)("p",null,"Let's send this into Compass so that it would be discoverable."),(0,r.kt)("p",null,"As of now, Compass supports ingesting assets via ",(0,r.kt)("inlineCode",{parentName:"p"},"gRPC")," and ",(0,r.kt)("inlineCode",{parentName:"p"},"http"),". In this example, we will use ",(0,r.kt)("inlineCode",{parentName:"p"},"http")," to send your first asset to Compass.\nCompass exposes an API ",(0,r.kt)("inlineCode",{parentName:"p"},"[PATCH] /v1beta1/assets")," to upload your asset."),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-bash"},'curl --location --request PATCH \'http://localhost:8080/v1beta1/assets\' \\\n--header \'Content-Type: application/json\' \\\n--header \'Compass-User-UUID: john.doe@example.com\' \\\n--data-raw \'{\n    "asset": {\n        "urn": "main-postgres:my-database.orders",\n        "type": "table",\n        "service": "postgres",\n        "name": "orders",\n        "data": {\n            "database": "my-database",\n            "namespace": "main-postgres"\n        }\n    }\n}\'\n')),(0,r.kt)("p",null,"There are a few things to notice here:"),(0,r.kt)("ol",null,(0,r.kt)("li",{parentName:"ol"},(0,r.kt)("p",{parentName:"li"},"The HTTP method used is ",(0,r.kt)("inlineCode",{parentName:"p"},"PATCH"),". This is because Compass does not have a dedicated ",(0,r.kt)("inlineCode",{parentName:"p"},"Create")," API, it uses a single API to ",(0,r.kt)("inlineCode",{parentName:"p"},"Patch / Create")," an asset instead. So when updating or patching your asset, you can use the same API.")),(0,r.kt)("li",{parentName:"ol"},(0,r.kt)("p",{parentName:"li"},"Compass requires ",(0,r.kt)("inlineCode",{parentName:"p"},"Compass-User-UUID")," header to be in the request. More information about the identity header can be found ",(0,r.kt)("a",{parentName:"p",href:"../concepts/user"},"here"),". To simplify this tour, let's just use ",(0,r.kt)("inlineCode",{parentName:"p"},"john.doe@example.com"),".")),(0,r.kt)("li",{parentName:"ol"},(0,r.kt)("p",{parentName:"li"},"When sending our asset to Compass, we need to put our asset object inside an ",(0,r.kt)("inlineCode",{parentName:"p"},"asset")," field as shown in the sample curl above."))),(0,r.kt)("p",null,"On a success insertion, your will receive below response:"),(0,r.kt)("pre",null,(0,r.kt)("code",{parentName:"pre",className:"language-json"},'{ "id": "cebeb793-8933-434c-b38f-beb6dbad91a5" }\n')),(0,r.kt)("p",null,(0,r.kt)("strong",{parentName:"p"},"id")," is an identifier of your asset. Unlike ",(0,r.kt)("inlineCode",{parentName:"p"},"urn")," which is provided by you, ",(0,r.kt)("inlineCode",{parentName:"p"},"id")," is auto generated by Compass if there was no asset found with the given URN."),(0,r.kt)("h2",{id:"conclusion"},"Conclusion"),(0,r.kt)("p",null,"Now that you have successfully ingested your asset to Compass, we can now search and find it via Compass."),(0,r.kt)("p",null,"In the next section, we will see how Compass can help you in searching and discovering your assets."))}d.isMDXComponent=!0}}]);