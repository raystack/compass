"use strict";(self.webpackChunkcompass=self.webpackChunkcompass||[]).push([[980],{3905:function(e,t,r){r.d(t,{Zo:function(){return l},kt:function(){return g}});var n=r(7294);function a(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function s(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,n)}return r}function o(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?s(Object(r),!0).forEach((function(t){a(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):s(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function i(e,t){if(null==e)return{};var r,n,a=function(e,t){if(null==e)return{};var r,n,a={},s=Object.keys(e);for(n=0;n<s.length;n++)r=s[n],t.indexOf(r)>=0||(a[r]=e[r]);return a}(e,t);if(Object.getOwnPropertySymbols){var s=Object.getOwnPropertySymbols(e);for(n=0;n<s.length;n++)r=s[n],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(a[r]=e[r])}return a}var c=n.createContext({}),u=function(e){var t=n.useContext(c),r=t;return e&&(r="function"==typeof e?e(t):o(o({},t),e)),r},l=function(e){var t=u(e.components);return n.createElement(c.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return n.createElement(n.Fragment,{},t)}},d=n.forwardRef((function(e,t){var r=e.components,a=e.mdxType,s=e.originalType,c=e.parentName,l=i(e,["components","mdxType","originalType","parentName"]),d=u(r),g=a,m=d["".concat(c,".").concat(g)]||d[g]||p[g]||s;return r?n.createElement(m,o(o({ref:t},l),{},{components:r})):n.createElement(m,o({ref:t},l))}));function g(e,t){var r=arguments,a=t&&t.mdxType;if("string"==typeof e||a){var s=r.length,o=new Array(s);o[0]=d;var i={};for(var c in t)hasOwnProperty.call(t,c)&&(i[c]=t[c]);i.originalType=e,i.mdxType="string"==typeof e?e:a,o[1]=i;for(var u=2;u<s;u++)o[u]=r[u];return n.createElement.apply(null,o)}return n.createElement.apply(null,r)}d.displayName="MDXCreateElement"},6484:function(e,t,r){r.r(t),r.d(t,{assets:function(){return l},contentTitle:function(){return c},default:function(){return g},frontMatter:function(){return i},metadata:function(){return u},toc:function(){return p}});var n=r(7462),a=r(3366),s=(r(7294),r(3905)),o=["components"],i={},c="Starring",u={unversionedId:"guides/starring",id:"guides/starring",title:"Starring",description:"Compass allows a user to stars an asset. This bookmarking functionality is introduced to increase the speed of a user to get information.",source:"@site/docs/guides/starring.md",sourceDirName:"guides",slug:"/guides/starring",permalink:"/compass/guides/starring",draft:!1,editUrl:"https://github.com/raystack/compass/edit/master/docs/docs/guides/starring.md",tags:[],version:"current",frontMatter:{},sidebar:"docsSidebar",previous:{title:"Querying metadata",permalink:"/compass/guides/querying"},next:{title:"Tagging",permalink:"/compass/guides/tagging"}},l={},p=[],d={toc:p};function g(e){var t=e.components,r=(0,a.Z)(e,o);return(0,s.kt)("wrapper",(0,n.Z)({},d,r,{components:t,mdxType:"MDXLayout"}),(0,s.kt)("h1",{id:"starring"},"Starring"),(0,s.kt)("p",null,"Compass allows a user to stars an asset. This bookmarking functionality is introduced to increase the speed of a user to get information."),(0,s.kt)("p",null,"To star and asset, we can use the User Starring API. Assuming we already have ",(0,s.kt)("inlineCode",{parentName:"p"},"asset_id")," that we want to star."),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},"$ curl --request PUT 'http://localhost:8080/v1beta1/me/starred/00c06ef7-badb-4236-9d9e-889697cbda46' \\\n--header 'Compass-User-UUID:raystack@email.com'\n")),(0,s.kt)("p",null,"To get the list of my starred assets."),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl --request PUT \'http://localhost:8080/v1beta1/me/starred\' \\\n--header \'Compass-User-UUID:raystack@email.com\'\n\n{\n  "data": [\n      {\n          "id": "00c06ef7-badb-4236-9d9e-889697cbda46",\n          "urn": "kafka::g-godata-id-playground/ g-godata-id-seg-enriched-booking-dagger",\n          "type": "topic",\n          "service": "kafka",\n          "name": "g-godata-id-seg-enriched-booking-dagger",\n          "description": "",\n          "labels": {\n              "flink_name": "g-godata-id-playground",\n              "sink_type": "kafka"\n          }\n      }\n  ]\n}\n')),(0,s.kt)("p",null,"There is also an API to see which users star an asset (stargazers) in the Asset API."),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl \'http://localhost:8080/v1beta1/assets/00c06ef7-badb-4236-9d9e-889697cbda46/stargazers\' \\\n--header \'Compass-User-UUID:raystack@email.com\'\n\n{\n  "data": [\n      {\n          "id": "1111-2222-3333",\n          "email": "raystack@email.com",\n          "provider": "shield"\n      }\n  ]\n}\n')))}g.isMDXComponent=!0}}]);