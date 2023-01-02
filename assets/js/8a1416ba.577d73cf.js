"use strict";(self.webpackChunkcompass=self.webpackChunkcompass||[]).push([[886],{3905:function(e,t,r){r.d(t,{Zo:function(){return p},kt:function(){return f}});var n=r(7294);function a(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function c(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,n)}return r}function s(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?c(Object(r),!0).forEach((function(t){a(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):c(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function o(e,t){if(null==e)return{};var r,n,a=function(e,t){if(null==e)return{};var r,n,a={},c=Object.keys(e);for(n=0;n<c.length;n++)r=c[n],t.indexOf(r)>=0||(a[r]=e[r]);return a}(e,t);if(Object.getOwnPropertySymbols){var c=Object.getOwnPropertySymbols(e);for(n=0;n<c.length;n++)r=c[n],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(a[r]=e[r])}return a}var i=n.createContext({}),l=function(e){var t=n.useContext(i),r=t;return e&&(r="function"==typeof e?e(t):s(s({},t),e)),r},p=function(e){var t=l(e.components);return n.createElement(i.Provider,{value:t},e.children)},u={inlineCode:"code",wrapper:function(e){var t=e.children;return n.createElement(n.Fragment,{},t)}},m=n.forwardRef((function(e,t){var r=e.components,a=e.mdxType,c=e.originalType,i=e.parentName,p=o(e,["components","mdxType","originalType","parentName"]),m=l(r),f=a,d=m["".concat(i,".").concat(f)]||m[f]||u[f]||c;return r?n.createElement(d,s(s({ref:t},p),{},{components:r})):n.createElement(d,s({ref:t},p))}));function f(e,t){var r=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var c=r.length,s=new Array(c);s[0]=m;var o={};for(var i in t)hasOwnProperty.call(t,i)&&(o[i]=t[i]);o.originalType=e,o.mdxType="string"==typeof e?e:a,s[1]=o;for(var l=2;l<c;l++)s[l]=r[l];return n.createElement.apply(null,s)}return n.createElement.apply(null,r)}m.displayName="MDXCreateElement"},4730:function(e,t,r){r.r(t),r.d(t,{assets:function(){return p},contentTitle:function(){return i},default:function(){return f},frontMatter:function(){return o},metadata:function(){return l},toc:function(){return u}});var n=r(7462),a=r(3366),c=(r(7294),r(3905)),s=["components"],o={},i="Architecture",l={unversionedId:"concepts/architecture",id:"concepts/architecture",title:"Architecture",description:"Compass' architecture is pretty simple. It has a client-server architecture backed by PostgreSQL as a main storage and Elasticsearch as a secondary storage and provides HTTP & gRPC interface to interact with.",source:"@site/docs/concepts/architecture.md",sourceDirName:"concepts",slug:"/concepts/architecture",permalink:"/compass/concepts/architecture",draft:!1,editUrl:"https://github.com/odpf/compass/edit/master/docs/docs/concepts/architecture.md",tags:[],version:"current",frontMatter:{},sidebar:"docsSidebar",previous:{title:"User",permalink:"/compass/concepts/user"},next:{title:"Internals",permalink:"/compass/concepts/internals"}},p={},u=[{value:"System Design",id:"system-design",level:2},{value:"Components",id:"components",level:3},{value:"gRPC Server",id:"grpc-server",level:4},{value:"gRPC-gateway Server",id:"grpc-gateway-server",level:4},{value:"PostgreSQL",id:"postgresql",level:4},{value:"Elasticsearch",id:"elasticsearch",level:4}],m={toc:u};function f(e){var t=e.components,o=(0,a.Z)(e,s);return(0,c.kt)("wrapper",(0,n.Z)({},m,o,{components:t,mdxType:"MDXLayout"}),(0,c.kt)("h1",{id:"architecture"},"Architecture"),(0,c.kt)("p",null,"Compass' architecture is pretty simple. It has a client-server architecture backed by PostgreSQL as a main storage and Elasticsearch as a secondary storage and provides HTTP & gRPC interface to interact with."),(0,c.kt)("p",null,(0,c.kt)("img",{alt:"Compass Architecture",src:r(8406).Z,width:"514",height:"291"})),(0,c.kt)("h2",{id:"system-design"},"System Design"),(0,c.kt)("h3",{id:"components"},"Components"),(0,c.kt)("h4",{id:"grpc-server"},"gRPC Server"),(0,c.kt)("ul",null,(0,c.kt)("li",{parentName:"ul"},"gRPC server is the main interface to interact with Compass."),(0,c.kt)("li",{parentName:"ul"},"The protobuf file to define the interface is centralized in ",(0,c.kt)("a",{parentName:"li",href:"https://github.com/odpf/proton/tree/main/odpf/compass/v1beta1"},"odpf/proton"))),(0,c.kt)("h4",{id:"grpc-gateway-server"},"gRPC-gateway Server"),(0,c.kt)("ul",null,(0,c.kt)("li",{parentName:"ul"},"gRPC-gateway server transcodes HTTP call to gRPC call and allows client to interact with Compass using RESTful HTTP request.")),(0,c.kt)("h4",{id:"postgresql"},"PostgreSQL"),(0,c.kt)("ul",null,(0,c.kt)("li",{parentName:"ul"},"Compass uses PostgreSQL as it is main storage for storing all of its metadata.")),(0,c.kt)("h4",{id:"elasticsearch"},"Elasticsearch"),(0,c.kt)("ul",null,(0,c.kt)("li",{parentName:"ul"},"Compass uses Elasticsearch as it is secondary storage to power search of metadata.")))}f.isMDXComponent=!0},8406:function(e,t,r){t.Z=r.p+"assets/images/architecture-c98ef0681046da635ca3b1302cf8b0f4.png"}}]);