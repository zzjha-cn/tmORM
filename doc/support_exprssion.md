## 查询操作符支持情况

### 已支持查询表达式
| 名称 | 说明 |
| --- | --- |
| `$eq` | 匹配等于指定值的文档 |
| `$gt` | 匹配大于指定值的文档 |
| `$gte` | 匹配大于等于指定值的文档 |
| `$in` | 匹配数组中任意一个值的文档 |
| `$lt` | 匹配小于指定值的文档 |
| `$lte` | 匹配小于等于指定值的文档 |
| `$ne` | 匹配不等于指定值的文档 |
| `$nin` | 匹配不在数组中的值的文档 |
|  |  |
| `$and` | 使用逻辑 `AND` 连接查询子句，返回与所有子句条件匹配的文档 |
| `$or` | 使用逻辑 `OR` 连接多个查询子句，返回符合任一子句条件的文档 |
|  |  |
| `$exists` | 匹配包含或不包含指定字段的文档 |
| `$type` | 匹配指定数据类型的文档 |
|  |  |
| `$expr` | 在查询中使用聚合表达式 |
| `$regex` | 正则表达式匹配 |
| `$mod` | 匹配除以指定数后有指定余数的值 |
| `$all` | 匹配数组中包含所有指定值的文档 |
|  |  |
| `$elemMatch` | 匹配数组中至少一个元素满足条件的文档 |
| `$size` | 匹配数组大小等于指定值的文档 |
|  |  |

### 未封装支持的查询表达式
| 名称 | 说明 |
| --- | --- |
| `$not` | 反转查询谓词的效果，返回与查询谓词不匹配的文档 |
| `$nor` | 使用逻辑 `NOR` 连接查询子句，返回无法匹配任一子句的文档 |
|  |  |
| `$jsonSchema` | 根据 JSON 模式验证文档 |
| `$text` | 文本搜索查询 |
| `$where` | 使用 JavaScript 表达式进行查询 |
|  |  |
| `$geoIntersects` | 匹配与指定几何对象相交的几何对象 |
| `$geoWithin` | 匹配位于指定几何对象内的几何对象 |
| `$near` | 按距离排序，匹配接近指定点的几何对象 |
| `$nearSphere` | 按球面距离排序，匹配接近指定点的几何对象 |
|  |  |
| `$bitsAllClear` | 指定的位全部为 0 |
| `$bitsAllSet` | 指定的位全部为 1 |
| `$bitsAnyClear` | 指定的位中至少有一个为 0 |
| `$bitsAnySet` | 指定的位中至少有一个为 1 |
|  |  |
| `$meta` | 预测在 `$text` 操作中分配的文件分数 |
| `$slice` | 限制从数组中投影的元素数量，支持跳过切片和对切片进行数量限制 |


### 已封装支持的聚合表达式
| 操作符 | 说明 | 示例 |
|--------|------|------|
| `$abs` | 返回一个数字的绝对值。 | `{ $abs: [ "$field" ] }` |
| `$add` | 添加数字以返回总和，或添加数字和日期以返回新日期。 | `{ $add: [ "$field1", "$field2" ] }` |
| `$divide` | 返回第一个数字除以第二个数字的结果。 | `{ $divide: [ "$field1", "$field2" ] }` |
| `$floor` | 返回小于或等于指定数字的最大整数。 | `{ $floor: [ "$field" ] }` |
| `$mod` | 返回第一个数字除以第二个数字的余数。 | `{ $mod: [ "$field1", "$field2" ] }` |
| `$multiply` | 将数字相乘以返回乘积。 | `{ $multiply: [ "$field1", "$field2" ] }` |
| `$subtract` | 返回第一个值减去第二个值后的结果。 | `{ $subtract: [ "$field1", "$field2" ] }` |
| | |  |
| `$arrayElemAt` | 返回数组中指定位置的元素。 | `{ $arrayElemAt: [ "$array", 0 ] }` |
| `$arrayToObject` | 将数组转换为对象。 | `{ $arrayToObject: [ "$array" ] }` |
| `$concatArrays` | 连接多个数组。 | `{ $concatArrays: [ "$array1", "$array2" ] }` |
| `$reverseArray` | 反转数组。 | `{ $reverseArray: [ "$array" ] }` |
| `$size` | 返回数组的大小。 | `{ $size: [ "$array" ] }` |
| `$slice` | 返回数组的子集。 | `{ $slice: [ "$array", 0, 5 ] }` |
| | |  |
| `$and` | 仅当 _所有_ 表达式的计算结果均为 `true` 时才返回 `true`。 | `{ $and: [ { $gt: [ "$field1", 10 ] }, { $lt: [ "$field2", 20 ] } ] }` |
| `$or` | 当 _任何_ 表达式的计算结果为 `true` 时，返回 `true`。 | `{ $or: [ { $gt: [ "$field1", 10 ] }, { $lt: [ "$field2", 20 ] } ] }` |
| | |  |
| `$eq` | 如果这些值相等，则返回 `true`。 | `{ $eq: [ "$field1", "$field2" ] }` |
| `$gt` | 如果第一个值大于第二个值，则返回 `true`。 | `{ $gt: [ "$field1", "$field2" ] }` |
| `$gte` | 如果第一个值大于等于第二个值，则返回 `true`。 | `{ $gte: [ "$field1", "$field2" ] }` |
| `$lt` | 如果第一个值小于第二个值，则返回 `true`。 | `{ $lt: [ "$field1", "$field2" ] }` |
| `$lte` | 如果第一个值小于等于第二个值，则返回 `true`。 | `{ $lte: [ "$field1", "$field2" ] }` |
| `$ne` | 如果值 _不_ 相等，则返回 `true`。 | `{ $ne: [ "$field1", "$field2" ] }` |
| | |  |
| `$cond` | 一种三元运算符，它可用于计算一个表达式，并根据结果返回另外两个表达式之一的值。 | `{ $cond: { if: { $gt: [ "$field", 10 ] }, then: "high", else: "low" } }` |
| | |  |
| `$concat` | 连接多个字符串。 | `{ $concat: [ "$string1", "$string2" ] }` |
| `$type` | 返回值的类型。 | `{ $type: [ "$field" ] }` |
| | |  |
| `$sum` | 计算总和。 | `{ $sum: "$field" }` |
| `$avg` | 计算平均值。 | `{ $avg: "$field" }` |
| `$min` | 找到最小值。 | `{ $min: "$field" }` |
| `$max` | 找到最大值。 | `{ $max: "$field" }` |
| `$first` | 返回第一个值。 | `{ $first: "$field" }` |
| `$last` | 返回最后一个值。 | `{ $last: "$field" }` |
| `$push` | 将值添加到数组。 | `{ $push: "$field" }` |
| `$addToSet` | 将唯一值添加到数组。 | `{ $addToSet: "$field" }` |
| `$count` | 计算文档数量。 | `{ $count: "count" }` |

### 未封装支持的聚合表达式操作符

| 操作符 | 说明 | 示例 |
|--------|------|------|
| `$log` | 以指定基数计算数字的对数。 | `{ $log: [ "$field", 10 ] }` |
| `$log10` | 计算一个数字以 10 为底的对数。 | `{ $log10: [ "$field" ] }` |
| `$trunc` | 将数字截断为整数或指定的小数位。 | `{ $trunc: [ "$field", 2 ] }` |
| | |  |
| `$pow` | 将一个数字提升到指定的指数。 | `{ $pow: [ "$field", 2 ] }` |
| `$round` | 将数字舍入到整数或指定的小数位。 | `{ $round: [ "$field", 2 ] }` |
| `$sqrt` | 计算平方根。 | `{ $sqrt: [ "$field" ] }` |
| `$bitAnd` | 返回对 `int` 或 `long` 值的数组执行按位 `and` 操作的结果。 | `{ $bitAnd: [ "$field1", "$field2" ] }` |
| `$bitNot` | 返回对单个参数或包含单个 `int` 或 `long` 值的数组执行按位 `not` 操作的结果。 | `{ $bitNot: [ "$field" ] }` |
| `$bitOr` | 返回对 `int` 或 `long` 值的数组执行按位 `or` 操作的结果。 | `{ $bitOr: [ "$field1", "$field2" ] }` |
| `$bitXor` | 返回对 `int` 和 `long` 值的数组执行按位 `xor`（排他或）操作的结果。 | `{ $bitXor: [ "$field1", "$field2" ] }` |
| | |  |
| `$filter` | 筛选数组中的元素。 | `{ $filter: { input: "$array", as: "item", cond: { $gt: [ "$$item", 10 ] } } }` |
| `$indexOfArray` | 返回数组中元素的索引。 | `{ $indexOfArray: [ "$array", "$value" ] }` |
| `$isArray` | 检查值是否为数组。 | `{ $isArray: [ "$value" ] }` |
| `$map` | 对数组中的每个元素应用表达式。 | `{ $map: { input: "$array", as: "item", in: { $add: [ "$$item", 1 ] } } }` |
| `$objectToArray` | 将对象转换为数组。 | `{ $objectToArray: [ "$object" ] }` |
| `$zip` | 将多个数组合并为一个数组。 | `{ $zip: { inputs: [ "$array1", "$array2" ] } }` |
| `$range` | 生成一个数字范围的数组。 | `{ $range: [ 0, 10 ] }` |
| `$reduce` | 对数组中的元素进行累积操作。 | `{ $reduce: { input: "$array", initialValue: 0, in: { $add: [ "$$value", "$$this" ] } } }` |
| `$not` | 返回与参数表达式相反的布尔值。 | `{ $not: [ { $gt: [ "$field", 10 ] } ] }` |
| | |  |
| `$cmp` | 如果两个值相等，则返回 `0`；如果第一个值大于第二个值，则返回 `1`；如果第一个值小于第二个值，则返回 `-1`。 | `{ $cmp: [ "$field1", "$field2" ] }` |
| | |  |
| `$ifNull` | 返回第一个表达式的非空结果；或者，如果第一个表达式生成空结果，则返回第二个表达式的结果。 | `{ $ifNull: [ "$field1", "$field2" ] }` |
| `$switch` | 对一系列 case 表达式求值。当它找到计算结果为 `true` 的表达式时， `$switch` 会执行指定表达式并脱离控制流。 | `{ $switch: { branches: [ { case: { $eq: [ "$field", "value" ] }, then: "result" } ], default: "default" } }` |
| | |  |
| `$dateAdd` | 将指定的时间间隔添加到日期。 | `{ $dateAdd: { startDate: "$date", unit: "day", amount: 1 } }` |
| `$dateFromParts` | 从部分构建日期。 | `{ $dateFromParts: { year: 2020, month: 1, day: 1 } }` |
| `$dateFromString` | 将字符串转换为日期。 | `{ $dateFromString: { dateString: "$dateString" } }` |
| `$dateSubtract` | 从日期中减去指定的时间间隔。 | `{ $dateSubtract: { startDate: "$date", unit: "day", amount: 1 } }` |
| `$dateToParts` | 将日期分解为部分。 | `{ $dateToParts: { date: "$date" } }` |
| `$dateToString` | 将日期转换为字符串。 | `{ $dateToString: { format: "%Y-%m-%d", date: "$date" } }` |
| `$dayOfMonth` | 返回日期的月份中的天数。 | `{ $dayOfMonth: [ "$date" ] }` |
| `$dayOfWeek` | 返回日期的星期几。 | `{ $dayOfWeek: [ "$date" ] }` |
| `$dayOfYear` | 返回日期的年份中的天数。 | `{ $dayOfYear: [ "$date" ] }` |
| `$hour` | 返回日期的小时。 | `{ $hour: [ "$date" ] }` |
| `$isoDayOfWeek` | 返回日期的 ISO 星期几。 | `{ $isoDayOfWeek: [ "$date" ] }` |
| `$isoWeek` | 返回日期的 ISO 周数。 | `{ $isoWeek: [ "$date" ] }` |
| `$isoYear` | 返回日期的 ISO 年份。 | `{ $isoYear: [ "$date" ] }` |
| `$millisecond` | 返回日期的毫秒。 | `{ $millisecond: [ "$date" ] }` |
| `$minute` | 返回日期的分钟。 | `{ $minute: [ "$date" ] }` |
| `$month` | 返回日期的月份。 | `{ $month: [ "$date" ] }` |
| `$second` | 返回日期的秒数。 | `{ $second: [ "$date" ] }` |
| `$week` | 返回日期的周数。 | `{ $week: [ "$date" ] }` |
| `$year` | 返回日期的年份。 | `{ $year: [ "$date" ] }` |
| | |  |
| `$indexOfBytes` | 返回子字符串在字符串中的索引（字节）。 | `{ $indexOfBytes: [ "$string", "substring" ] }` |
| `$indexOfCP` | 返回子字符串在字符串中的索引（代码点）。 | `{ $indexOfCP: [ "$string", "substring" ] }` |
| `$ltrim` | 去除字符串左侧的空格。 | `{ $ltrim: [ "$string" ] }` |
| `$rtrim` | 去除字符串右侧的空格。 | `{ $rtrim: [ "$string" ] }` |
| `$split` | 将字符串按指定分隔符分割为数组。 | `{ $split: [ "$string", "," ] }` |
| `$strLenBytes` | 返回字符串的长度（字节）。 | `{ $strLenBytes: [ "$string" ] }` |
| `$strLenCP` | 返回字符串的长度（代码点）。 | `{ $strLenCP: [ "$string" ] }` |
| `$strcasecmp` | 比较两个字符串（不区分大小写）。 | `{ $strcasecmp: [ "$string1", "$string2" ] }` |
| `$substr` | 提取字符串的子字符串。 | `{ $substr: [ "$string", 0, 5 ] }` |
| `$substrBytes` | 提取字符串的子字符串（字节）。 | `{ $substrBytes: [ "$string", 0, 5 ] }` |
| `$substrCP` | 提取字符串的子字符串（代码点）。 | `{ $substrCP: [ "$string", 0, 5 ] }` |
| `$toLower` | 将字符串转换为小写。 | `{ $toLower: [ "$string" ] }` |
| `$toUpper` | 将字符串转换为大写。 | `{ $toUpper: [ "$string" ] }` |
| `$trim` | 去除字符串两侧的空格。 | `{ $trim: [ "$string" ] }` |
| | |  |
| `$convert` | 将值从一种类型转换为另一种类型。 | `{ $convert: { input: "$field", to: "string" } }` |
| `$toString` | 将值转换为字符串。 | `{ $toString: [ "$field" ] }` |
| `$toInt` | 将值转换为整数。 | `{ $toInt: [ "$field" ] }` |
| `$toDouble` | 将值转换为双精度浮点数。 | `{ $toDouble: [ "$field" ] }` |
| `$toDecimal` | 将值转换为十进制。 | `{ $toDecimal: [ "$field" ] }` |
| `$toLong` | 将值转换为长整数。 | `{ $toLong: [ "$field" ] }` |
| `$toDate` | 将值转换为日期。 | `{ $toDate: [ "$field" ] }` |
| `$toObjectId` | 将值转换为 ObjectId。 | `{ $toObjectId: [ "$field" ] }` |
| | |  |
| `$denseRank` | 返回密集排名。 | `{ $denseRank: {} }` |
| `$documentNumber` | 返回文档在集合中的位置。 | `{ $documentNumber: {} }` |
| `$rank` | 返回排名。 | `{ $rank: {} }` |
| `$rowNumber` | 返回行号。 | `{ $rowNumber: {} }` |
| | |  |
| `$literal` | 返回一个值而不进行解析。 | `{ $literal: [ "$value" ] }` |
| `$meta` | 访问与聚合操作相关的每个文档的可用元数据。 | `{ $meta: "textScore" }` |
