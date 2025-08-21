const http = require('http');
const fs = require('fs');
const path = require('path');

// 测试图片上传功能
async function testImageUpload() {
    console.log('开始测试图片处理功能...\n');

    // 测试 1: 健康检查
    console.log('1. 测试健康检查端点');
    try {
        const response = await fetch('http://localhost:8080/health');
        const data = await response.json();
        console.log('✅ 健康检查:', data);
    } catch (error) {
        console.log('❌ 健康检查失败:', error.message);
        console.log('请确保服务器正在运行: go run cmd/main.go');
        return;
    }

    // 测试 2: 测试图片上传端点存在性
    console.log('\n2. 测试图片上传端点');
    try {
        const response = await fetch('http://localhost:8080/api/v1/images/upload', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({})
        });
        
        if (response.status === 400) {
            console.log('✅ 图片上传端点存在（返回400是因为没有文件）');
        } else {
            console.log('📝 图片上传端点状态:', response.status);
        }
    } catch (error) {
        console.log('❌ 图片上传端点测试失败:', error.message);
    }

    // 测试 3: 测试图片列表端点
    console.log('\n3. 测试图片列表端点');
    try {
        const response = await fetch('http://localhost:8080/api/v1/images/');
        const data = await response.json();
        console.log('✅ 图片列表端点:', data);
    } catch (error) {
        console.log('❌ 图片列表端点测试失败:', error.message);
    }

    console.log('\n测试完成！');
    console.log('\n使用方法:');
    console.log('1. 确保R2存储凭证已配置在.env文件中');
    console.log('2. 使用以下curl命令测试文件上传:');
    console.log(`
curl -X POST http://localhost:8080/api/v1/images/upload \\
  -F "file=@/path/to/your/image.jpg" \\
  -F "is_public=true" \\
  -F "folder=test"
    `);
    console.log('3. 查看上传的图片:');
    console.log('GET http://localhost:8080/api/v1/images/');
    console.log('\n4. API端点列表:');
    console.log('- POST /api/v1/images/upload - 上传图片');
    console.log('- GET  /api/v1/images/      - 获取图片列表');
    console.log('- GET  /api/v1/images/:id   - 获取图片详情');
    console.log('- GET  /api/v1/images/public/:key - 获取公开图片');
    console.log('- PUT  /api/v1/images/:id/refresh - 刷新图片URL');
    console.log('- DELETE /api/v1/images/:id - 删除图片');
}

// 如果没有fetch，使用简单的node.js实现
if (typeof fetch === 'undefined') {
    global.fetch = async function(url, options = {}) {
        const { URL } = require('url');
        const urlObj = new URL(url);
        
        const requestOptions = {
            hostname: urlObj.hostname,
            port: urlObj.port || (urlObj.protocol === 'https:' ? 443 : 80),
            path: urlObj.pathname + urlObj.search,
            method: options.method || 'GET',
            headers: options.headers || {}
        };

        return new Promise((resolve, reject) => {
            const req = http.request(requestOptions, (res) => {
                let data = '';
                res.on('data', (chunk) => {
                    data += chunk;
                });
                res.on('end', () => {
                    resolve({
                        status: res.statusCode,
                        json: () => Promise.resolve(JSON.parse(data)),
                        text: () => Promise.resolve(data)
                    });
                });
            });

            req.on('error', reject);

            if (options.body) {
                req.write(options.body);
            }

            req.end();
        });
    };
}

testImageUpload().catch(console.error);