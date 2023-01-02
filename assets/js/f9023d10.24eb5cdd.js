"use strict";(self.webpackChunkcompass=self.webpackChunkcompass||[]).push([[19],{3905:function(e,n,s){s.d(n,{Zo:function(){return d},kt:function(){return m}});var t=s(7294);function i(e,n,s){return n in e?Object.defineProperty(e,n,{value:s,enumerable:!0,configurable:!0,writable:!0}):e[n]=s,e}function a(e,n){var s=Object.keys(e);if(Object.getOwnPropertySymbols){var t=Object.getOwnPropertySymbols(e);n&&(t=t.filter((function(n){return Object.getOwnPropertyDescriptor(e,n).enumerable}))),s.push.apply(s,t)}return s}function o(e){for(var n=1;n<arguments.length;n++){var s=null!=arguments[n]?arguments[n]:{};n%2?a(Object(s),!0).forEach((function(n){i(e,n,s[n])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(s)):a(Object(s)).forEach((function(n){Object.defineProperty(e,n,Object.getOwnPropertyDescriptor(s,n))}))}return e}function r(e,n){if(null==e)return{};var s,t,i=function(e,n){if(null==e)return{};var s,t,i={},a=Object.keys(e);for(t=0;t<a.length;t++)s=a[t],n.indexOf(s)>=0||(i[s]=e[s]);return i}(e,n);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(t=0;t<a.length;t++)s=a[t],n.indexOf(s)>=0||Object.prototype.propertyIsEnumerable.call(e,s)&&(i[s]=e[s])}return i}var l=t.createContext({}),c=function(e){var n=t.useContext(l),s=n;return e&&(s="function"==typeof e?e(n):o(o({},n),e)),s},d=function(e){var n=c(e.components);return t.createElement(l.Provider,{value:n},e.children)},u={inlineCode:"code",wrapper:function(e){var n=e.children;return t.createElement(t.Fragment,{},n)}},p=t.forwardRef((function(e,n){var s=e.components,i=e.mdxType,a=e.originalType,l=e.parentName,d=r(e,["components","mdxType","originalType","parentName"]),p=c(s),m=i,h=p["".concat(l,".").concat(m)]||p[m]||u[m]||a;return s?t.createElement(h,o(o({ref:n},d),{},{components:s})):t.createElement(h,o({ref:n},d))}));function m(e,n){var s=arguments,i=n&&n.mdxType;if("string"==typeof e||i){var a=s.length,o=new Array(a);o[0]=p;var r={};for(var l in n)hasOwnProperty.call(n,l)&&(r[l]=n[l]);r.originalType=e,r.mdxType="string"==typeof e?e:i,o[1]=r;for(var c=2;c<a;c++)o[c]=s[c];return t.createElement.apply(null,o)}return t.createElement.apply(null,s)}p.displayName="MDXCreateElement"},6936:function(e,n,s){s.r(n),s.d(n,{assets:function(){return d},contentTitle:function(){return l},default:function(){return m},frontMatter:function(){return r},metadata:function(){return c},toc:function(){return u}});var t=s(7462),i=s(3366),a=(s(7294),s(3905)),o=["components"],r={},l="Discussion",c={unversionedId:"guides/discussion",id:"guides/discussion",title:"Discussion",description:"Discussion is a new feature in Compass. One could create a discussion and all users can put comment in it. Currently, there are three types of discussions issues, open ended, and question and answer. Depending on the type, the discussion could have multiple possible states. In the current version, all types only have two states: open and closed. A newly created discussion will always be assign an open state.",source:"@site/docs/guides/discussion.md",sourceDirName:"guides",slug:"/guides/discussion",permalink:"/compass/guides/discussion",draft:!1,editUrl:"https://github.com/odpf/compass/edit/master/docs/docs/guides/discussion.md",tags:[],version:"current",frontMatter:{},sidebar:"docsSidebar",previous:{title:"Tagging",permalink:"/compass/guides/tagging"},next:{title:"Overview",permalink:"/compass/concepts/overview"}},d={},u=[{value:"Create a Discussion",id:"create-a-discussion",level:2},{value:"Fetching All Discussions",id:"fetching-all-discussions",level:2},{value:"Patching Discussion",id:"patching-discussion",level:2},{value:"Commenting a Discussion",id:"commenting-a-discussion",level:2},{value:"Getting All My Discussions",id:"getting-all-my-discussions",level:2}],p={toc:u};function m(e){var n=e.components,s=(0,i.Z)(e,o);return(0,a.kt)("wrapper",(0,t.Z)({},p,s,{components:n,mdxType:"MDXLayout"}),(0,a.kt)("h1",{id:"discussion"},"Discussion"),(0,a.kt)("p",null,"Discussion is a new feature in Compass. One could create a discussion and all users can put comment in it. Currently, there are three types of discussions ",(0,a.kt)("inlineCode",{parentName:"p"},"issues"),", ",(0,a.kt)("inlineCode",{parentName:"p"},"open ended"),", and ",(0,a.kt)("inlineCode",{parentName:"p"},"question and answer"),". Depending on the type, the discussion could have multiple possible states. In the current version, all types only have two states: ",(0,a.kt)("inlineCode",{parentName:"p"},"open")," and ",(0,a.kt)("inlineCode",{parentName:"p"},"closed"),". A newly created discussion will always be assign an ",(0,a.kt)("inlineCode",{parentName:"p"},"open")," state."),(0,a.kt)("h2",{id:"create-a-discussion"},"Create a Discussion"),(0,a.kt)("p",null,"A discussion thread can be created with the Discussion API. The API contract is available ",(0,a.kt)("a",{parentName:"p",href:"https://github.com/odpf/compass/blob/main/third_party/OpenAPI/compass.swagger.json"},"here"),"."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl --request POST \'http://localhost:8080/v1beta1/discussions\' \\\n--header \'Compass-User-UUID:odpf@email.com\' \\\n--data-raw \'{\n  "title": "The first discussion",\n  "body": "This is the first discussion thread in Compass",\n  "type": "openended"\n}\'\n')),(0,a.kt)("h2",{id:"fetching-all-discussions"},"Fetching All Discussions"),(0,a.kt)("p",null,"The Get Discussions will fetch all discussions in Compass."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"$ curl 'http://localhost:8080/v1beta1/discussions' \\\n--header 'Compass-User-UUID:odpf@email.com'\n")),(0,a.kt)("p",null,"The response will be something like"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-javascript"},'{\n    "data": [\n        {\n            "id": "1",\n            "title": "The first discussion",\n            "body": "This is the first discussion thread in Compass",\n            "type": "openended"\n            "state": "open",\n            "labels": [],\n            "assets": [],\n            "assignees": [],\n            "owner": {\n                "id": "dd9e2e07-a13f-1c2b-07e3-e32cf0f7688c",\n                "email": "odpf@email.com",\n                "provider": "shield"\n            },\n            "created_at": "elit cillum Duis",\n            "updated_at": "velit dolor ex"\n        }\n    ]\n}\n')),(0,a.kt)("p",null,"Notice the state is ",(0,a.kt)("inlineCode",{parentName:"p"},"open")," by default once we create a new discussion. There are also some additional features in discussion where we can label the discussion and assign users and assets to the discussion. These labelling and assinging assets and users could also be done when we are creating a discussion."),(0,a.kt)("h2",{id:"patching-discussion"},"Patching Discussion"),(0,a.kt)("p",null,"If we are not labelling and assigning users & assets to the discussion in the creation step, there are also a dedicated API to do those."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl --request PATCH \'http://localhost:8080/v1beta1/discussions/1\' \\\n--header \'Compass-User-UUID:odpf@email.com\' \\\n--data-raw \'{\n    "title": "The first discussion (duplicated)",\n    "state": "closed"\n}\'\n')),(0,a.kt)("p",null,"We just need to send the fields that we want to patch for a discussion. Some fields have array type, in this case the PATCH will overwrite the fields with the new value."),(0,a.kt)("p",null,"For example we have this labelled discussion."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl \'http://localhost:8080/v1beta1/discussions\' \\\n--header \'Compass-User-UUID:odpf@email.com\'\n\n{\n    "data": [\n        {\n            "id": "1",\n            "title": "The first discussion",\n            "body": "This is the first discussion thread in Compass",\n            "type": "openended"\n            "state": "open",\n            "labels": [\n                "work",\n                "urgent",\n                "help wanted"\n            ],\n            "owner": {\n                "id": "dd9e2e07-a13f-1c2b-07e3-e32cf0f7688c",\n                "email": "odpf@email.com",\n                "provider": "shield"\n            },\n            "created_at": "elit cillum Duis",\n            "updated_at": "velit dolor ex"\n        }\n    ]\n}\n')),(0,a.kt)("p",null,"If we patch the label with the new values."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"$ curl --request PATCH 'http://localhost:8080/v1beta1/discussions/1' \\\n--header 'Compass-User-UUID:odpf@email.com' \\\n--data-raw '{\n    \"labels\": [\"new value\"]\n}'\n")),(0,a.kt)("p",null,"The discussion with id 1 will be updated like this."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl \'http://localhost:8080/v1beta1/discussions\' \\\n--header \'Compass-User-UUID:odpf@email.com\'\n\n{\n    "data": [\n        {\n            "id": "1",\n            "title": "The first discussion",\n            "body": "This is the first discussion thread in Compass",\n            "type": "openended"\n            "state": "open",\n            "labels": [\n                "new value"\n            ],\n            "owner": {\n                "id": "dd9e2e07-a13f-1c2b-07e3-e32cf0f7688c",\n                "email": "odpf@email.com",\n                "provider": "shield"\n            },\n            "created_at": "elit cillum Duis",\n            "updated_at": "velit dolor ex"\n        }\n    ]\n}\n')),(0,a.kt)("h2",{id:"commenting-a-discussion"},"Commenting a Discussion"),(0,a.kt)("p",null,"One could also comment a specific discussion with discussion comment API."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},"$ curl --request POST 'http://localhost:8080/v1beta1/discussions/1/comments' \\\n--header 'Compass-User-UUID:odpf@email.com' \\\n--data-raw '{\n  \"body\": \"This is the first comment of discussion 1\"\n}'\n")),(0,a.kt)("h2",{id:"getting-all-my-discussions"},"Getting All My Discussions"),(0,a.kt)("p",null,"Compass integrates discussions with User API so we could fetch all discussions belong to us with this API."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-bash"},'$ curl \'http://localhost:8080/v1beta1/me/discussions\' \\\n--header \'Compass-User-UUID:odpf@email.com\'\n\n{\n    "data": [\n        {\n            "id": "1",\n            "title": "The first discussion",\n            "body": "This is the first discussion thread in Compass",\n            "type": "openended"\n            "state": "open",\n            "labels": [\n                "new value"\n            ],\n            "owner": {\n                "id": "dd9e2e07-a13f-1c2b-07e3-e32cf0f7688c",\n                "email": "odpf@email.com",\n                "provider": "shield"\n            },\n            "created_at": "elit cillum Duis",\n            "updated_at": "velit dolor ex"\n        }\n    ]\n}\n')))}m.isMDXComponent=!0}}]);